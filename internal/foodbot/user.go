package foodbot

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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

// RespondTo the given message from the user.
func (u *User) RespondTo(msg string) (string, error) {
	if msg == "/start" {
		u.state = AskedForLimit
		return "Hi\\! What's your daily limit \\(kcal\\)?", nil
	}

	if msg == "/cancel" {
		return u.handleCancel()
	}

	switch u.state {
	case AskedForLimit:
		return u.handleLimit(msg)
	case AskedForProduct, AskedForKcal, AskedForGrams:
		return u.handleAdd(msg)
	}

	switch msg {
	case "/limit":
		u.state = AskedForLimit
		return "Ok, what's your new daily limit \\(kcal\\)?", nil
	case "/add":
		u.inProgress.When = time.Now()
		u.state = AskedForProduct
		return "All right\\! Tell me, what have you eaten?", nil
	case "/stat":
		return formatStat(u.todayReports(), u.Limit), nil
	case "/stat7":
		return formatStat7(u.weeklyStat(), u.Limit), nil
	}

	return "", errors.New("I don't understand you")
}

// weeklyStat for this user.
func (u *User) weeklyStat() []dayResult {
	var week []string
	now := time.Now()
	for delta := 0; delta <= 6; delta++ {
		week = append(week, now.AddDate(0, 0, -delta).Format("Mon 2006/01/02"))
	}

	history := bot.History(u.ID, week...)
	var ret []dayResult
	for _, day := range week {
		total := history[day]
		ret = append(ret, dayResult{
			Date:    day,
			Kcal:    total,
			InLimit: total < u.Limit,
		})
	}
	return ret
}

// todayReports returns food eaten by this user today.
func (u *User) todayReports() []sqlite.Report {
	reports := bot.TodayReports(u.ID)
	sort.Slice(reports, func(i, j int) bool { return reports[i].When.Before(reports[j].When) })
	return reports
}

func (u *User) handleCancel() (string, error) {
	switch u.state {
	case Default:
		return "Nothing to cancel\\.\\.\\. Maybe /add food or see /stat for today?", nil
	case AskedForLimit:
		u.state = Default
		return fmt.Sprintf("Ok\\. Your limit is still %v kcal\\.", u.Limit), nil
	default:
		u.inProgress = sqlite.Report{}
		u.state = Default
		return "All right, no food has been reported\\.", nil
	}
}

func (u *User) handleLimit(msg string) (string, error) {
	limit, err := strconv.ParseUint(msg, 10, 32)
	if err != nil {
		return "", fmt.Errorf("%q is not an integer. Enter your daily limit (kcal)", msg)
	}

	u.Limit = uint32(limit)
	u.state = Default
	return "Limit saved, thanks\\! Now you can /add food or see /stat for today\\.", nil
}

func (u *User) handleAdd(msg string) (string, error) {
	switch u.state {
	case AskedForProduct:
		u.inProgress.Product = msg
		u.state = AskedForKcal

		kcals, ok := bot.GetProductKcals(u.inProgress.Product)
		if !ok {
			u.state = AskedForKcal
			return fmt.Sprintf("How many calories \\(kcal per 💯g\\) are there in `%q`?", u.inProgress.Product), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Choose kcal for `%q` from the list:\n", u.inProgress.Product)
		for _, kcal := range kcals {
			fmt.Fprintf(&sb, "*/%v kcal*\n", kcal)
		}
		fmt.Fprintf(&sb, "\nOr enter new calorie amount \\(kcal per 💯g\\)\\.\n")

		return sb.String(), nil

	case AskedForKcal:
		var kcalReg = regexp.MustCompile(`/[0-9]+`)
		if kcalReg.MatchString(msg) {
			msg = msg[1:]
		}

		kcal, err := strconv.ParseUint(msg, 10, 32)
		if err != nil {
			return "", fmt.Errorf("%q is not an integer. Enter kcal per 💯g for %q", msg, u.inProgress.Product)
		}

		u.inProgress.Kcal = uint32(kcal)
		u.state = AskedForGrams
		return fmt.Sprintf("How many grams of `%q` have you eaten?", u.inProgress.Product), nil

	case AskedForGrams:
		grams, err := strconv.ParseUint(msg, 10, 32)
		if err != nil {
			return "", fmt.Errorf("%q is not an integer. Enter how many grams you've eaten", msg)
		}

		u.inProgress.Grams = uint32(grams)
		bot.SaveReport(u.ID, u.inProgress)
		bot.AddProductKcal(u.inProgress.Product, u.inProgress.Kcal)

		total := TotalKcal(bot.TodayReports(u.ID))
		var ret string
		if total < u.Limit {
			ret = fmt.Sprintf("Noted\\. *%v kcal* left for today 😋\nLet's /add more food\\.", u.Limit-total)
		} else {
			ret = fmt.Sprintf("Noted\\. You ate *%v kcal* over the limit 😱\nYou can see /stat7 for the last week\\.", total-u.Limit)
		}

		u.inProgress = sqlite.Report{}
		u.state = Default
		return ret, nil
	default:
		return "", nil
	}
}
