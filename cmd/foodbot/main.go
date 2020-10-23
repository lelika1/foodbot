package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lelika1/foodbot/internal/foodbot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s bot_token", os.Args[0])
	}

	db := foodbot.NewDB()

	bot, err := tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't understand you")
		switch update.Message.Text {
		case "/add":
			msg.Text = "Add some food"
		case "/stat":
			msg.Text = todayReportResponse(update.Message.From.UserName, db)
			msg.ParseMode = "MarkdownV2"
		case "/stat7":
			msg.Text = weeklyReport(update.Message.From.UserName, db)
			msg.ParseMode = "MarkdownV2"
		default:
			msg.ReplyToMessageID = update.Message.MessageID
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		bot.Send(msg)
	}
}

func weeklyReport(username string, db *foodbot.DB) string {
	report, err := db.WeeklyReport(username)
	switch err {
	case nil:
		return report
	case foodbot.ErrUserNotFound:
		return "You aren't a user of this bot\\."
	}

	log.Printf("WeeklyReport(%v) failed with %v", username, err)
	return "Something went wrong\\. Try later\\."
}

func todayReportResponse(username string, db *foodbot.DB) string {
	reports, err := db.TodayReports(username)
	switch err {
	case nil:
		break
	case foodbot.ErrUserNotFound:
		return "You aren't a user of this bot\\."
	default:
		log.Printf("TodayReport(%v) failed with %v", username, err)
		return "Something went wrong. Try later\\."
	}

	if len(reports) == 0 {
		return "*You ate nothing so far\\.*"
	}

	preFormat := make([]struct {
		begin  string
		end    string
		len    int
		spaces int
	}, len(reports))

	var total uint32
	maxLen := 0
	for i, r := range reports {
		preFormat[i].begin = fmt.Sprintf("%s: %v", r.Time.Format("15:04:05"), r.Product)
		kcal := r.Kcal * r.Grams / 100
		preFormat[i].end = fmt.Sprintf("%v kcal", kcal)

		preFormat[i].len = len(preFormat[i].begin) + len(preFormat[i].end)
		if maxLen < preFormat[i].len {
			maxLen = preFormat[i].len
		}

		total += kcal
	}

	for i := range preFormat {
		preFormat[i].spaces = maxLen - preFormat[i].len + 1
	}

	var sb strings.Builder
	sb.WriteString("*You ate today:*\n")
	for _, f := range preFormat {
		fmt.Fprintf(&sb, "`%s%s`*%s*\n", f.begin, strings.Repeat(" ", f.spaces), f.end)
	}
	fmt.Fprintf(&sb, "\n`Total:` *%v kcal*\n", total)
	return sb.String()
}
