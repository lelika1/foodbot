package foodbot

import (
	"sort"
	"time"
)

// User of the bot.
type User struct {
	Name    string
	Limit   uint32
	History map[string]uint32 // "2006/01/02" -> kcal consumed
	Today   Day
	State   State
}

// State of the communication with the user.
type State uint8

// All possible states of the user.
const (
	Default State = iota
	AskedForLimit
)

// NewUser creates a new user.
func NewUser(name string, limit uint32) *User {
	return &User{
		Name:    name,
		Limit:   limit,
		History: make(map[string]uint32),
	}
}

// RespondTo the given message from the user.
func (u *User) RespondTo(msg string) string {
	switch msg {
	case "/start":
		return "Input daily limit"
	case "/add":
		return "Add some food"
	case "/stat":
		return formatDayReport(u.todayReports())
	case "/stat7":
		return formatWeeklyReport(u.weeklyReport())
	}
	return "I don't understand you\\."
}

// weeklyReport for this user.
func (u *User) weeklyReport() weeklyReport {
	now := time.Now()
	var history []shortDayReport
	for delta := 1; delta <= 6; delta++ {
		key := now.AddDate(0, 0, -delta).Format("Mon 2006/01/02")
		kcal := u.History[key]
		history = append(history, shortDayReport{
			Date:    key,
			Kcal:    kcal,
			InLimit: kcal < u.Limit,
		})
	}

	total := u.Today.TotalKcal()
	return weeklyReport{
		Today:        total,
		TodayInLimit: total < u.Limit,
		History:      history,
	}
}

// todayReports returns food eaten by this user today.
func (u *User) todayReports() []Report {
	reports := u.Today.Reports
	sort.Slice(reports, func(i, j int) bool { return reports[i].Time.Before(reports[j].Time) })
	return reports
}

// AddFood consumed by this user.
func (u *User) AddFood(product string, kcal, grams uint32) {
	u.Today.Reports = append(u.Today.Reports, Report{
		Time:    time.Now(),
		Product: product,
		Kcal:    kcal,
		Grams:   grams,
	})
}
