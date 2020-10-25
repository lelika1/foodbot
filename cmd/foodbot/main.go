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
	if err == foodbot.ErrUserNotFound {
		return "You aren't a user of this bot\\."
	}

	if err != nil {
		log.Printf("WeeklyReport(%v) failed with %v", username, err)
		return "Something went wrong\\. Try later\\."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "`%s Today:         ` *%v kcal*\n", color(report.TodayInLimit), report.Today)
	for _, r := range report.History {
		fmt.Fprintf(&sb, "`%s %v:` *%v kcal*\n", color(r.InLimit), r.Date, r.Kcal)
	}
	return sb.String()
}

func todayReportResponse(username string, db *foodbot.DB) string {
	reports, err := db.TodayReports(username)
	if err == foodbot.ErrUserNotFound {
		return "You aren't a user of this bot\\."
	}

	if err != nil {
		log.Printf("TodayReport(%v) failed with %v", username, err)
		return "Something went wrong. Try later\\."
	}

	if len(reports) == 0 {
		return "*You ate nothing so far\\.*"
	}

	type Line struct {
		Begin, End string
	}

	var lines []Line

	var total uint32
	var maxLen int
	for i, r := range reports {
		kcal := r.Kcal * r.Grams / 100
		total += kcal

		lines = append(lines, Line{
			Begin: fmt.Sprintf("%s: %v", r.Time.Format("15:04:05"), r.Product),
			End:   fmt.Sprintf("%v kcal", kcal),
		})

		if l := len(lines[i].Begin) + len(lines[i].End); maxLen < l {
			maxLen = l
		}
	}

	var sb strings.Builder
	sb.WriteString("*You ate today:*\n")
	for _, line := range lines {
		spaces := maxLen - (len(line.Begin) + len(line.End)) + 1
		fmt.Fprintf(&sb, "`%s%s`*%s*\n", line.Begin, strings.Repeat(" ", spaces), line.End)
	}
	fmt.Fprintf(&sb, "\n`Total:` *%v kcal*\n", total)
	return sb.String()
}

func color(inLimit bool) string {
	if inLimit {
		return "✅"
	}
	return "❌"
}
