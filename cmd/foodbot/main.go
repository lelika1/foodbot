package main

import (
	"log"
	"os"

	"github.com/lelika1/foodbot/internal/foodbot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s bot_token database_path", os.Args[0])
	}

	bot, err := tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	db := foodbot.NewDB(os.Args[2])

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't understand you")
		user, err := db.User(update.Message.From.UserName)
		switch err {
		case nil:
			if text, err := user.RespondTo(update.Message.Text); err != nil {
				msg.Text = err.Error()
				msg.ReplyToMessageID = update.Message.MessageID
			} else {
				msg.Text = text
				msg.ParseMode = "MarkdownV2"
			}

		case foodbot.ErrUserNotFound:
			msg.Text = "You aren't a user of this bot."
		default:
			msg.Text = err.Error()
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		bot.Send(msg)
	}
}
