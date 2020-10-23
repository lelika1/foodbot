package foodbot

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

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

// User ...
type User struct {
	Name    string
	Limit   uint32
	History map[string]uint32 // "2006/01/02" -> kcal consumed
	Today   Day
}

// NewUser ...
func NewUser(name string, limit uint32) *User {
	return &User{
		Name:    name,
		Limit:   limit,
		History: make(map[string]uint32),
	}
}

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

// TodayReports ...
func (db *DB) TodayReports(username string) ([]Report, error) {
	user, err := db.User(username)
	if err != nil {
		return nil, err
	}

	reports := user.Today.Reports
	sort.Slice(reports, func(i, j int) bool { return reports[i].Time.Before(reports[j].Time) })
	return reports, nil
}

// WeeklyReport ...
func (db *DB) WeeklyReport(username string) (string, error) {
	user, err := db.User(username)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	total := user.Today.TotalKcal()
	fmt.Fprintf(&sb, "`%s Today:         ` *%v kcal*\n", color(total, user.Limit), total)

	now := time.Now()
	for delta := 1; delta <= 6; delta++ {
		key := now.AddDate(0, 0, -delta).Format("Mon 2006/01/02")
		kcal := user.History[key]
		fmt.Fprintf(&sb, "`%s %v:` *%v kcal*\n", color(kcal, user.Limit), key, kcal)
	}
	return sb.String(), nil

}

func color(val, limit uint32) string {
	if val < limit {
		return "✅"
	}
	return "❌"
}
