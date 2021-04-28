# Foodbot

`Foodbot` is a [Telegram](http://telegram.org/) bot written in Go that helps you report and track your daily calories consumption.

## How to run Foodbot

In order to run `Foodbot` you have to [create a telegram bot](https://core.telegram.org/bots#3-how-do-i-create-a-bot) and get an `authorization token` for it.

Since we rely on SQLite, you also need to install binary dependencies for [mattn/go-sqlite3](http://github.com/mattn/go-sqlite3), as described in their README.

After that, you can simply `go run` to build and start `Foodbot`:

```bash
go run github.com/lelika1/foodbot/cmd/foodbot $TGBOT_AUTH_TOKEN $PATH_TO_DB
```

## Bot usage

When you start talking to the bot for the first time from a new Telegram account, it will ask you about your daily calories limit (it can later be changed using `/limit` command).

After this short registration phase you can use the following commands:

* `/add` - start a flow for reporting new food you ate.
  * The bot will ask the name of a particular dish. It suggests you a couple of dishes you reported most recently. It helps adding food faster.
  * For every dish Foodbot remembers all possible nutrition values (`calories/100g`) and suggest you to either pick an existing one, or enter a new one.
  * The last step - enter the amount of food you ate in grams.
* `/stat` - report statistics about what you ate today and how many calories you still have. 
* `/stat7` - a high-level report for the past 7 days to see which days you were in limits and which days you overate.
* `/limit` - change you current daily calorie limit
* `/cancel` - escape from a multi-staged command without commiting its results. 

## Implementation details

We use [go-telegram-bot-api/telegram-bot-api](http://github.com/go-telegram-bot-api/telegram-bot-api) library to implement a telegram bot. 

The [main loop](https://github.com/lelika1/foodbot/blob/main/cmd/foodbot/main.go#L35-L48) of the programm processes updates from the Telegram API one by one, updating data for the user and their in-memory state. We persist user state into SQLite to be able to handle program restarts.

Since we need to support multiple users concurrently talking to the bot as well as processing multi-stage commands (i.e. `/add`, during which the users answers several questions one after another), we have to maintain a state machine of the conversation explicitly. Relevant code is in [internal/foodbot/user.go](https://github.com/lelika1/foodbot/blob/main/internal/foodbot/user.go).

`Foodbot` can store data in any SQL database since we work against a standard [`sql.DB`](https://pkg.go.dev/database/sql#DB) interface.

For simplicity, we register [mattn/go-sqlite3](http://github.com/mattn/go-sqlite3) SQL driver and store data in a SQLite database file on disc. This can be abstracted easily changed though.

