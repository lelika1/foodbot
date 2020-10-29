package foodbot

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// DB stores all the information of the bot.
type DB struct {
	Users map[string]*User
}

// NewDB loads DB from the given file, or creates a new one if file doesn't exist.
func NewDB(path string) *DB {
	// TODO use normal database
	LoadProducts("products.db")
	db := DB{Users: make(map[string]*User)}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	defer file.Close()
	if err != nil {
		log.Printf("os.OpenFile(%v) failed with %v\n", path, err)
		return &db
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		userDescr := strings.Split(line, " ")
		if len(userDescr) != 2 {
			log.Panic(fmt.Errorf("wrong line in the db: %q", line))
		}

		limit, err := strconv.ParseUint(userDescr[1], 10, 32)
		if err != nil {
			log.Panic(err)
		}

		db.Users[userDescr[0]] = NewUser(userDescr[0], uint32(limit))
	}
	return &db
}

// ErrUserNotFound means the bot doesn't have such user.
var ErrUserNotFound = errors.New("user was not found")

// User finds an user with the given name or returns an ErrUserNotFound if such user doesn't exists.
func (db *DB) User(name string) (*User, error) {
	if user, ok := db.Users[name]; ok {
		return user, nil
	}

	return nil, ErrUserNotFound
}

// Report ...
type Report struct {
	Time    time.Time
	Product string
	Kcal    uint32 // for 100g
	Grams   uint32
}

// Day ...
type Day struct {
	Reports []Report
}

// TotalKcal ...
func (d *Day) TotalKcal() uint32 {
	var ret uint32
	for _, r := range d.Reports {
		ret += r.Kcal * r.Grams / 100
	}
	return ret
}
