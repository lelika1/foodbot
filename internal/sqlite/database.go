package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver that supports database/sql.
)

// DB stores all bot data.
type DB struct {
	db *sql.DB
}

// NewDB opens the database and creates tables if they don't exist.
func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open db in file %q failed with %s", path, err)
	}

	_, err = db.Exec(createTablesQuery)
	if err != nil {
		return nil, fmt.Errorf("create tables for %q failed with %s", path, err)
	}

	return &DB{db: db}, nil
}

// Close connection to the DB.
func (d *DB) Close() {
	d.db.Close()
}

// Users of the bot.
func (d *DB) Users() []User {
	rows, err := d.db.Query("SELECT * FROM USER;")
	if err != nil {
		log.Printf("'SELECT * FROM USER;' failed with: %q", err)
		return nil
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err = rows.Scan(&u.ID, &u.Name, &u.Limit); err == nil {
			users = append(users, u)
		}
	}
	return users
}

// Products saved in the bot.
func (d *DB) Products() []Product {
	rows, err := d.db.Query("SELECT LOWER(NAME), KCAL FROM PRODUCT;")
	if err != nil {
		log.Printf("'SELECT LOWER(NAME), KCAL FROM PRODUCT;' failed with: %q", err)
		return nil
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err = rows.Scan(&p.Name, &p.Kcal); err == nil {
			products = append(products, p)
		}
	}
	return products
}

// TodayReports of the user.
func (d *DB) TodayReports(uid int) []Report {
	rows, err := d.db.Query(selectTodayQuery, uid, time.Now().Format("Mon 2006/01/02"))
	if err != nil {
		log.Printf("%q failed with: %q", selectTodayQuery, err)
		return nil
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var r Report
		var date, hours string
		if err = rows.Scan(&date, &hours, &r.Product, &r.Kcal, &r.Grams); err == nil {
			r.When, _ = time.Parse("Mon 2006/01/02 15:04:05", date+" "+hours)
			reports = append(reports, r)
		}
	}
	return reports
}

// History of the user in the given days.
func (d *DB) History(uid int, dates ...string) map[string]uint32 {
	sql, args := selectReportsQuery(uid, dates...)
	rows, err := d.db.Query(sql, args...)
	if err != nil {
		log.Printf("%q failed with: %q", sql, err)
		return nil
	}
	defer rows.Close()

	history := make(map[string]uint32)
	for rows.Next() {
		var date string
		var kcal uint32
		if err = rows.Scan(&date, &kcal); err == nil {
			history[date] = kcal
		}
	}
	return history
}

// SaveProduct into the database.
func (d *DB) SaveProduct(food string, kcal uint32) {
	if stmt, err := d.db.Prepare(insertProductQuery); err == nil {
		if _, err := stmt.Exec(strings.ToLower(food), kcal); err != nil {
			log.Printf("Exec failed with: %q", err)
		}
	} else {
		log.Printf("Prepare failed with: %q", err)
	}
}

// SaveReport of the user into the database.
func (d *DB) SaveReport(uid int, r Report) {
	if stmt, err := d.db.Prepare(insertReportQuery); err == nil {
		date := r.When.Format("Mon 2006/01/02")
		hours := r.When.Format("15:04:05")
		if _, err := stmt.Exec(uid, date, hours, r.Product, r.Kcal, r.Grams); err != nil {
			log.Printf("Exec failed with: %q", err)
		}
	} else {
		log.Printf("Prepare failed with: %q", err)
	}
}
