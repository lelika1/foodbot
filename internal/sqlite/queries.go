package sqlite

import "strings"

const createTablesQuery = `
    CREATE TABLE IF NOT EXISTS USER(
		ID              INTEGER   PRIMARY KEY   AUTOINCREMENT NOT NULL,
		NAME            TEXT                                  NOT NULL,
		DAILY_LIMIT           INTEGER   DEFAULT 0                   NOT NULL,
		UNIQUE (NAME)
	);
	CREATE TABLE IF NOT EXISTS PRODUCT(
        NAME            TEXT                                  NOT NULL,
        KCAL            INTEGER   DEFAULT 0                   NOT NULL,
        UNIQUE (NAME, KCAL)
	);
	CREATE TABLE IF NOT EXISTS REPORTS(
		USER_ID         INTEGER                               NOT NULL,
		DATE            TEXT                                  NOT NULL,
		TIME            TEXT                                  NOT NULL,
		PRODUCT         TEXT                                  NOT NULL,
        KCAL            INTEGER                               NOT NULL,
        GRAMS           INTEGER                               NOT NULL,
		FOREIGN KEY(USER_ID) REFERENCES USER(ID) ON DELETE CASCADE
	);`

const insertProductQuery = `INSERT OR IGNORE INTO PRODUCT(name, kcal) values(?, ?);`

const insertReportQuery = `INSERT INTO REPORTS(user_id, date, time, product, kcal, grams) values(?, ?, ?, ?, ?, ?);`

const selectTodayQuery = "SELECT DATE, TIME, LOWER(PRODUCT), KCAL, GRAMS FROM REPORTS WHERE USER_ID=? AND DATE=? AND GRAMS!=0;"

func selectReportsQuery(uid int, dates ...string) (string, []interface{}) {
	if len(dates) == 0 {
		return "SELECT DATE, SUM(KCAL * GRAMS)/100 FROM REPORTS WHERE DATE IN() GROUP BY DATE;", []interface{}{uid}
	}

	var sb strings.Builder
	sb.WriteString("SELECT DATE, SUM(KCAL * GRAMS)/100  FROM REPORTS WHERE DATE IN(?")
	sb.WriteString(strings.Repeat(",?", len(dates)-1))
	sb.WriteString(") GROUP BY DATE;")

	args := []interface{}{uid}
	for _, date := range dates {
		args = append(args, date)
	}
	return sb.String(), args
}
