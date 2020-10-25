package foodbot

import (
	"errors"
	"sort"
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

// AddFood ...
func (d *Day) AddFood(product string, kcal uint32, grams uint32) {
	d.Reports = append(d.Reports, Report{
		Time:    time.Now(),
		Product: product,
		Kcal:    kcal,
		Grams:   grams,
	})
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

// AddFood ...
func (db *DB) AddFood(username, product string, kcal, grams uint32) error {
	user, err := db.User(username)
	if err != nil {
		return err
	}

	user.Today.AddFood(product, kcal, grams)
	return nil
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

// ShortDayReport ...
type ShortDayReport struct {
	Date    string
	Kcal    uint32
	InLimit bool
}

// WeeklyReport ...
type WeeklyReport struct {
	Today        uint32
	TodayInLimit bool
	History      []ShortDayReport
}

// WeeklyReport ...
func (db *DB) WeeklyReport(username string) (WeeklyReport, error) {
	user, err := db.User(username)
	if err != nil {
		return WeeklyReport{}, err
	}

	now := time.Now()
	var history []ShortDayReport
	for delta := 1; delta <= 6; delta++ {
		key := now.AddDate(0, 0, -delta).Format("Mon 2006/01/02")
		kcal := user.History[key]
		history = append(history, ShortDayReport{
			Date:    key,
			Kcal:    kcal,
			InLimit: kcal < user.Limit,
		})
	}

	total := user.Today.TotalKcal()
	return WeeklyReport{
		Today:        total,
		TodayInLimit: total < user.Limit,
		History:      history,
	}, nil
}
