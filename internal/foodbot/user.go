package foodbot

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"
)

// User of the bot.
type User struct {
	Name    string
	Limit   uint32
	History map[string]uint32 // "2006/01/02" -> kcal consumed
	Today   Day
	State   State

	last Report
}

// State of the communication with the user.
type State uint8

// All possible states of the user.
const (
	Default State = iota
	AskedForLimit
	AskedForProduct
	AskedForKcal
	AskedForGrams
)

// NewUser creates a new user.
func NewUser(name string, limit uint32) *User {
	return &User{
		Name:    name,
		Limit:   limit,
		History: make(map[string]uint32),
		State:   Default,
	}
}

// RespondTo the given message from the user.
func (u *User) RespondTo(msg string) (string, error) {
	if u.State == AskedForLimit {
		limit, err := strconv.ParseUint(msg, 10, 32)
		if err != nil {
			return "", fmt.Errorf("%q is not a number. Input daily limit", msg)
		}

		u.Limit = uint32(limit)
		u.State = Default
		return "Limit was saved\\. Thanks\\!", nil
	}

	if u.State == AskedForProduct {
		u.last.Product = msg
		u.State = AskedForKcal
		return fmt.Sprintf("What is the calorie content of `%q`? Input kcal per 100g:", msg), nil
	}

	if u.State == AskedForKcal {
		kcal, err := strconv.ParseUint(msg, 10, 32)
		if err != nil {
			return "", fmt.Errorf("%q is not a number. Input kcal per 100g", msg)
		}

		u.last.Kcal = uint32(kcal)
		u.State = AskedForGrams
		return fmt.Sprintf("How many grams of `%q` have you eaten?", u.last.Product), nil
	}

	if u.State == AskedForGrams {
		grams, err := strconv.ParseUint(msg, 10, 32)
		if err != nil {
			return "", fmt.Errorf("%q is not a number. Input how many grams you've eaten", msg)
		}

		u.last.Grams = uint32(grams)

		u.State = Default
		u.Today.Reports = append(u.Today.Reports, u.last)
		ret := fmt.Sprintf("%v grams of `%q` with %v kcal for 100g was saved\\. Thanks\\!", u.last.Grams, u.last.Product, u.last.Kcal)
		u.last = Report{}
		return ret, nil
	}

	switch msg {
	case "/start":
		u.State = AskedForLimit
		return "Input daily limit:", nil
	case "/limit":
		u.State = AskedForLimit
		return "Input new daily limit:", nil
	case "/add":
		u.last.Time = time.Now()
		u.State = AskedForProduct
		return "What product do you want to report? Input product's name", nil
	case "/stat":
		return formatDayReport(u.todayReports()), nil
	case "/stat7":
		return formatWeeklyReport(u.weeklyReport()), nil
	}

	return "", errors.New("I don't understand you")
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
