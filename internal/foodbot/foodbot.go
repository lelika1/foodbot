package foodbot

import (
	"errors"

	"github.com/lelika1/foodbot/internal/sqlite"
)

// Bot stores all the information and connection to the database.
type Bot struct {
	db       *sqlite.DB
	users    map[string]*user
	products products
}

// NewBot connects to the given DB and loads all stored information for the foodbot.
func NewBot(dbPath string) (*Bot, error) {
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		return nil, err
	}

	return &Bot{
		db:       db,
		products: newProducts(db.Products()),
		users:    createUsers(db.Users()),
	}, nil
}

// Stop connection to the database.
func (b *Bot) Stop() {
	b.db.Close()
}

// errUserNotFound means the bot doesn't have such user.
var errUserNotFound = errors.New("user was not found")

// User finds an user with the given name.
// Returns an ErrUserNotFound if such user doesn't exists.
func (b *Bot) user(name string) (*user, error) {
	if user, ok := b.users[name]; ok {
		return user, nil
	}

	return nil, errUserNotFound
}

// AddProduct adds a new energy value for the given product.
func (b *Bot) AddProduct(name string, kcal uint32) {
	if b.products.add(name, kcal) {
		b.db.SaveProduct(name, kcal)
	}
}
