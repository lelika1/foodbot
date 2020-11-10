package foodbot

import (
	"errors"

	"github.com/lelika1/foodbot/internal/sqlite"
)

// Bot stores all the information and connection to the database.
type Bot struct {
	*sqlite.DB

	users    map[string]*User
	products Products
}

// NewBot connects to the given DB and loads all stored information for the foodbot.
func NewBot(dbPath string) (*Bot, error) {
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		return nil, err
	}

	return &Bot{
		users:    createUsers(db.Users()),
		products: createProducts(db.Products()),
		DB:       db,
	}, nil
}

// Stop connection to the database.
func (b *Bot) Stop() {
	b.Close()
}

// ErrUserNotFound means the bot doesn't have such user.
var ErrUserNotFound = errors.New("user was not found")

// User finds an user with the given name.
// Returns an ErrUserNotFound if such user doesn't exists.
func (b *Bot) User(name string) (*User, error) {
	if user, ok := b.users[name]; ok {
		return user, nil
	}

	return nil, ErrUserNotFound
}

// TotalKcal calculates energy of all eaten food.
func TotalKcal(reports []sqlite.Report) uint32 {
	var ret uint32
	for _, r := range reports {
		ret += r.Kcal * r.Grams
	}
	return ret / 100
}
