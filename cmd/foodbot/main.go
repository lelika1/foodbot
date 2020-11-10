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

	fbot, err := foodbot.NewBot(os.Args[2])
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't understand you")
		switch text, err := fbot.RespondTo(update.Message.From.UserName, update.Message.Text); err {
		case nil:
			msg.Text = text
			msg.ParseMode = "MarkdownV2"
		case foodbot.ErrUserNotFound:
			msg.Text = "You aren't a user of this bot."
			msg.ReplyToMessageID = update.Message.MessageID
		default:
			msg.Text = err.Error()
			msg.ReplyToMessageID = update.Message.MessageID
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		bot.Send(msg)
	}
}
