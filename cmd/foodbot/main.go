package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s bot_token", os.Args[0])
	}

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
			msg.Text = fmt.Sprintf("You ate since %v", time.Now().AddDate(0, 0, -6).Format("2006/01/02"))
		default:
			msg.ReplyToMessageID = update.Message.MessageID
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		bot.Send(msg)
	}
}
