package foodbot

// SQLCreateTables ...
const SQLCreateTables = `
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
	CREATE TABLE IF NOT EXISTS TODAY(
		USER_ID         INTEGER                               NOT NULL,
		TIME            TEXT                                  NOT NULL,
		PRODUCT         TEXT                                  NOT NULL,
        KCAL            INTEGER                               NOT NULL,
        GRAMS           INTEGER                               NOT NULL,
		FOREIGN KEY(USER_ID) REFERENCES USER(ID) ON DELETE CASCADE
	);`

//SQLInsertProduct ...
const SQLInsertProduct = `"INSERT INTO PRODUCT(name, kcal) values(?,?)"`

// SQLInsertTodayReport ...
const SQLInsertTodayReport = `INSERT INTO TODAY(user_id, time, product, kcal, grams) values(?, ?, ?, ?, ?);`
