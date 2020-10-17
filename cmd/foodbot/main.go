package main

import (
	"log"
	"os"

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
			msg.Text = "You ate today:"
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
		return "You aren't a user of this bot."
	}

	log.Printf("WeeklyReport(%v) failed with %v", username, err)
	return "Something went wrong. Try later."
}
