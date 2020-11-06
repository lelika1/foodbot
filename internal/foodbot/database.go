package foodbot

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver that supports database/sql.
)

// SQLDb is a wrap for a sql database.
type SQLDb struct {
	db *sql.DB
}

// ConnectSQLDb and creates tables if they don't exist.
func ConnectSQLDb(path string) (*SQLDb, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open db in file %q failed with %s", path, err)
	}

	_, err = db.Exec(SQLCreateTables)
	if err != nil {
		return nil, fmt.Errorf("create tables for %q failed with %s", path, err)
	}

	return &SQLDb{db: db}, nil
}

// CloseConnection to the DB.
func (sdb *SQLDb) CloseConnection() {
	//TODO save all information before close.
	sdb.db.Close()
}

// ErrUserNotFound means the bot doesn't have such user.
var ErrUserNotFound = errors.New("user was not found")

// FindUserID for the user with the given name.
// Returns an ErrUserNotFound if such user doesn't exists.
func (sdb *SQLDb) FindUserID(name string) (int, error) {
	rows, err := sdb.db.Query("SELECT ID FROM USER WHERE NAME=?;", name)
	if err != nil || !rows.Next() {
		return 0, ErrUserNotFound
	}
	defer rows.Close()

	var id int
	if err := rows.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// LoadUsers of the bot.
func (sdb *SQLDb) LoadUsers() map[int]*User {
	users := make(map[int]*User)

	rows, err := sdb.db.Query("SELECT * FROM USER;")
	if err != nil {
		return users
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id    int
			name  string
			limit uint32
		)
		if err = rows.Scan(&id, &name, &limit); err == nil {
			users[id] = NewUser(name, limit, id)
		}
	}
	return users
}

// LoadProducts saved in the bot.
func (sdb *SQLDb) LoadProducts() Products {
	products := make(map[string]map[uint32]bool)
	rows, err := sdb.db.Query("SELECT * FROM PRODUCT;")
	if err != nil {
		log.Println(err)
		return products
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name string
			kcal uint32
		)
		if err = rows.Scan(&name, &kcal); err == nil {
			if _, ok := products[name]; !ok {
				products[name] = make(map[uint32]bool)
			}
			products[name][kcal] = true
		}
	}
	return products
}

// GetHistoryForDates as a map "Mon 2006/01/02" -> list of eated food.
func (sdb *SQLDb) GetHistoryForDates(uid int, dates []string) map[string][]Report {
	history := make(map[string][]Report)
	for _, d := range dates {
		history[d] = nil
	}

	rows, err := sdb.db.Query("SELECT TIME, PRODUCT, KCAL, GRAMS FROM TODAY where user_id=?;", uid)
	if err != nil {
		return history
	}
	defer rows.Close()

	for rows.Next() {
		var (
			date, product string
			kcal, grams   uint32
		)

		if err = rows.Scan(&date, &product, &kcal, &grams); err == nil {
			when, _ := time.Parse("Jan 2 15:04:05 2006", date)
			date := when.Format("Mon 2006/01/02")
			if _, ok := history[date]; ok {
				history[date] = append(history[date], Report{
					When:    when,
					Product: product,
					Kcal:    kcal,
					Grams:   grams,
				})
			}
		}
	}
	return history
}

func (sdb *SQLDb) insertProduct(food string, kcal uint32) {
	if stmt, err := sdb.db.Prepare(SQLInsertProduct); err == nil {
		_, err := stmt.Exec(food, kcal)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
}

func (sdb *SQLDb) insertReport(uid int, r Report) {
	if stmt, err := sdb.db.Prepare(SQLInsertTodayReport); err == nil {
		_, err := stmt.Exec(uid, r.When.Format("Jan 2 15:04:05 2006"), r.Product, r.Kcal, r.Grams)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
}
