package foodbot

import (
	"errors"
	"time"
)

// DB ...
type DB struct {
	Users map[string]*User
}

// NewDB ...
func NewDB() *DB {
	return &DB{Users: map[string]*User{
		"osycheva": NewUser("osycheva", 1546),
	}}
}

// ErrUserNotFound ...
var ErrUserNotFound = errors.New("user was not found")

// User ...
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
