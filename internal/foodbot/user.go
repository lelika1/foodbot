package foodbot

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lelika1/foodbot/internal/sqlite"
)

// User of the bot.
type User struct {
	sqlite.User

	state State
	// For storing incomplete report (during AskedFor{Product,Kcal,Grams} states).
	inProgress sqlite.Report
}

// State of the communication with the user.
type State uint8

// All possible states of the user.
const (
	Default State = iota
	AskedForLimit
	AskedForProduct
	AskedForKcal
	AskedForGrams
)

func createUsers(users []sqlite.User) map[string]*User {
	ret := make(map[string]*User)
	for _, u := range users {
		ret[u.Name] = &User{User: u, state: Default}
	}
	return ret
}

// RespondToKeyboard the given message from the user.
func (b *Bot) RespondToKeyboard(msg *tgbotapi.CallbackQuery) tgbotapi.MessageConfig {
	chatID := msg.Message.Chat.ID
	msgID := msg.Message.MessageID

	u, err := b.User(msg.From.UserName)
	if err != nil {
		return errResponse(chatID, msgID, "You aren't a user of this bot.")
	}

	var p sqlite.Product
	json.Unmarshal([]byte(msg.Data), &p)
	if u.state == AskedForProduct || u.state == AskedForKcal {
		u.inProgress.Product = p
		u.state = AskedForGrams
		return response(chatID, fmt.Sprintf("How many grams of `%q` have you eaten?", u.inProgress.Name), true)

	}
	return errResponse(chatID, msgID, "I don't understand you")
}

// RespondTo the given message from the user.
func (b *Bot) RespondTo(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	chatID := msg.Chat.ID
	msgID := msg.MessageID

	u, err := b.User(msg.From.UserName)
	if err != nil {
		return errResponse(chatID, msgID, "You aren't a user of this bot.")
	}

	input := msg.Text
	if input == "/start" {
		u.state = AskedForLimit
		return response(chatID, "Hi! What's your daily limit (kcal)?", false)
	}

	if input == "/cancel" {
		return b.handleCancel(u, chatID)
	}

	switch u.state {
	case AskedForLimit:
		return b.handleLimit(u, chatID, msgID, input)
	case AskedForProduct, AskedForKcal, AskedForGrams:
		return b.handleAdd(u, chatID, msgID, input)
	}

	switch input {
	case "/limit":
		u.state = AskedForLimit
		return response(chatID, "Ok, what's your new daily limit (kcal)?", false)
	case "/add":
		u.inProgress.When = time.Now()
		u.state = AskedForProduct
		last := b.LastProducts(5)
		if len(last) == 0 {
			return response(chatID, "All right! Tell me, what have you eaten?", false)
		}

		ret := tgbotapi.NewMessage(msg.Chat.ID, "")
		ret.Text = "These products were recently reported to the bot. Choose one of them, or enter what have you eaten.\n"
		var rows [][]tgbotapi.InlineKeyboardButton
		for _, p := range last {
			data, _ := json.Marshal(p)
			row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
				Product(p).String(), string(data)))
			rows = append(rows, row)
		}
		ret.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		return ret
	case "/stat":
		return response(chatID, formatStat(b.todayReports(u), u.Limit), true)
	case "/stat7":
		return response(chatID, formatStat7(b.weeklyStat(u), u.Limit), true)
	}

	return errResponse(chatID, msgID, "I don't understand you")
}

// weeklyStat for this user.
func (b *Bot) weeklyStat(u *User) []dayResult {
	var week []time.Time
	now := time.Now()
	for delta := 0; delta <= 6; delta++ {
		week = append(week, now.AddDate(0, 0, -delta))
	}

	history := b.History(u.ID, week...)
	var ret []dayResult
	for _, day := range week {
		date := day.Format("Mon 2006/01/02")
		total := history[date]
		ret = append(ret, dayResult{
			Date:    date,
			Kcal:    total,
			InLimit: total < u.Limit,
		})
	}
	return ret
}

// todayReports returns food eaten by this user today.
func (b *Bot) todayReports(u *User) []sqlite.Report {
	reports := b.TodayReports(u.ID)
	sort.Slice(reports, func(i, j int) bool { return reports[i].When.Before(reports[j].When) })
	return reports
}

func (b *Bot) handleCancel(u *User, chatID int64) tgbotapi.MessageConfig {
	switch u.state {
	case Default:
		return response(chatID, "Nothing to cancel... Maybe /add food or see /stat for today?", false)
	case AskedForLimit:
		u.state = Default
		return response(chatID, fmt.Sprintf("Ok. Your limit is still %v kcal.", u.Limit), false)
	default:
		u.inProgress = sqlite.Report{}
		u.state = Default
		return response(chatID, "All right, no food has been reported.", false)
	}
}

func (b *Bot) handleLimit(u *User, chatID int64, msgID int, text string) tgbotapi.MessageConfig {
	limit, err := strconv.ParseUint(text, 10, 32)
	if err != nil {
		return errResponse(chatID, msgID, fmt.Sprintf("%q is not an integer. Enter your daily limit (kcal)", text))
	}

	u.Limit = uint32(limit)
	u.state = Default
	return response(chatID, "Limit saved, thanks! Now you can /add food or see /stat for today.", false)
}

func (b *Bot) handleAdd(u *User, chatID int64, msgID int, text string) tgbotapi.MessageConfig {
	switch u.state {
	case AskedForProduct:
		u.inProgress.Name = text
		u.state = AskedForKcal

		products := b.GetProducts(u.inProgress.Name)

		if len(products) == 0 {
			u.state = AskedForKcal
			return response(chatID, fmt.Sprintf("How many calories \\(kcal per 💯g\\) are there in `%q`?", u.inProgress.Name), true)
		}

		ret := tgbotapi.NewMessage(chatID, "")
		ret.Text = fmt.Sprintf("Choose one of the products from the list or enter new calorie amount \\(kcal per 💯g\\) for %q\\.\n", u.inProgress.Name)
		ret.ParseMode = "MarkdownV2"
		var rows [][]tgbotapi.InlineKeyboardButton
		for _, p := range products {
			data, _ := json.Marshal(p)
			row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(p.String(), string(data)))
			rows = append(rows, row)
		}
		ret.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		return ret

	case AskedForKcal:
		kcal, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return errResponse(chatID, msgID, fmt.Sprintf("%q is not an integer. Enter kcal per 💯g for %q", text, u.inProgress.Name))
		}

		u.inProgress.Kcal = uint32(kcal)
		u.state = AskedForGrams
		return response(chatID, fmt.Sprintf("How many grams of `%q` have you eaten?", u.inProgress.Name), true)

	case AskedForGrams:
		grams, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return errResponse(chatID, msgID, fmt.Sprintf("%q is not an integer. Enter how many grams you've eaten", text))
		}

		u.inProgress.Grams = uint32(grams)
		b.SaveReport(u.ID, u.inProgress)
		b.AddProduct(u.inProgress.Name, u.inProgress.Kcal)

		total := TotalKcal(b.TodayReports(u.ID))
		var ret string
		if total < u.Limit {
			ret = fmt.Sprintf("Noted\\. *%v kcal* left for today 😋\nLet's /add more food\\.", u.Limit-total)
		} else {
			ret = fmt.Sprintf("Noted\\. You ate *%v kcal* over the limit 😱\nYou can see /stat7 for the last week\\.", total-u.Limit)
		}

		u.inProgress = sqlite.Report{}
		u.state = Default
		return response(chatID, ret, true)
	default:
		return response(chatID, "", false)
	}
}

func response(chatID int64, text string, isMarkdown bool) tgbotapi.MessageConfig {
	ret := tgbotapi.NewMessage(chatID, "")
	ret.Text = text
	if isMarkdown {
		ret.ParseMode = "MarkdownV2"
	}
	return ret
}

func errResponse(chatID int64, messageID int, text string) tgbotapi.MessageConfig {
	ret := tgbotapi.NewMessage(chatID, "")
	ret.Text = text
	ret.ReplyToMessageID = messageID
	return ret
}
