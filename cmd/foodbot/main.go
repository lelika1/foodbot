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
		if update.CallbackQuery != nil {
			msg := fbot.RespondToKeyboard(update.CallbackQuery)
			bot.Send(msg)
			continue
		}

		if update.Message != nil {
			msg := fbot.RespondTo(update.Message)
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			bot.Send(msg)
		}
	}
}
