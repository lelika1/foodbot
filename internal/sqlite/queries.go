package sqlite

import (
	"strings"
	"time"
)

const createTablesQuery = `
    CREATE TABLE IF NOT EXISTS USER(
		ID              INTEGER   PRIMARY KEY   AUTOINCREMENT NOT NULL,
		NAME            TEXT                                  NOT NULL,
		DAILY_LIMIT		INTEGER   DEFAULT 0                   NOT NULL,
		TIMEZONE		TEXT 	  DEFAULT 'UTC'				  NOT NULL,
		UNIQUE (NAME)
	);
	CREATE TABLE IF NOT EXISTS PRODUCT(
        NAME            TEXT                                  NOT NULL,
        KCAL            INTEGER   DEFAULT 0                   NOT NULL,
        UNIQUE (NAME, KCAL)
	);
	CREATE TABLE IF NOT EXISTS REPORTS(
		USER_ID         INTEGER                               NOT NULL,
		TIME			INTEGER								  NOT NULL,
		PRODUCT         TEXT                                  NOT NULL,
        KCAL            INTEGER                               NOT NULL,
        GRAMS           INTEGER                               NOT NULL,
		FOREIGN KEY(USER_ID) REFERENCES USER(ID) ON DELETE CASCADE
	);`

const insertProductQuery = `INSERT OR IGNORE INTO PRODUCT(name, kcal) values(?, ?);`

const insertReportQuery = `INSERT INTO REPORTS(user_id, time, product, kcal, grams) values(?, ?, ?, ?, ?);`

const selectTodayQuery = "SELECT TIME, LOWER(PRODUCT), KCAL, GRAMS FROM REPORTS WHERE USER_ID=?1 AND GRAMS!=0 AND (TIME + ?2)/?3 = ?4;"

const selectLastProducts = "SELECT MAX(TIME), LOWER(PRODUCT), KCAL FROM REPORTS group by LOWER(PRODUCT), KCAL order by time desc limit ?;"

const secondsInDay = 24 * 60 * 60

func selectReportsQuery(uid int, offset int64, dates ...time.Time) (string, []interface{}) {
	var sb strings.Builder
	sb.WriteString("SELECT TIME, SUM(KCAL * GRAMS)/100 FROM REPORTS WHERE USER_ID=?1 AND (TIME+?2)/?3 IN(")
	sb.WriteString(strings.Trim(strings.Repeat("?,", len(dates)), ","))
	sb.WriteString(") GROUP BY (TIME+?2)/?3;")

	args := []interface{}{uid, offset, secondsInDay}
	for _, date := range dates {
		args = append(args, (date.Unix()+offset)/secondsInDay)
	}
	return sb.String(), args
}
