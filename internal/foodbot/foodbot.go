package foodbot

import (
	"time"
)

var bot *Bot

// Bot stores all the information and connection to the database.
type Bot struct {
	Users    map[int]*User
	products Products

	db *SQLDb
}

// Products stores information about products as (name: {kcal1: true, kcal2: true}).
type Products map[string]map[uint32]bool

// NewBot connects to the given DB and loads all stored information for the foodbot.
func NewBot(dbPath string) (*Bot, error) {
	db, err := ConnectSQLDb(dbPath)
	if err != nil {
		return nil, err
	}
	users := db.LoadUsers()
	products := db.LoadProducts()
	bot = &Bot{Users: users, products: products, db: db}
	return bot, nil
}

// Stop connection to the database.
func (bot *Bot) Stop() {
	bot.db.CloseConnection()
}

// User finds an user with the given name.
// Returns an ErrUserNotFound if such user doesn't exists.
func (bot *Bot) User(name string) (*User, error) {
	id, err := bot.db.FindUserID(name)
	if err != nil {
		return nil, err
	}

	if user, ok := bot.Users[id]; ok {
		return user, nil
	}

	return nil, ErrUserNotFound
}

// Report ...
type Report struct {
	When    time.Time
	Product string
	Kcal    uint32 // for 100g
	Grams   uint32
}

// TotalKcal calculates energy of all eaten food.
func TotalKcal(reports []Report) uint32 {
	var ret uint32
	for _, r := range reports {
		ret += r.Kcal * r.Grams
	}
	return ret / 100
}
