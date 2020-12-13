package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DataSourceName string

func InitDB(dataSourceName string) (*sql.DB, error) { //
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return db, err
	}
	statement, err :=
		db.Prepare(`CREATE TABLE IF NOT EXISTS users (
			email VARCHAR(320) PRIMARY KEY, 
			hashedEmail CHAR(32), 
			hashedPassword CHAR(32),
			token CHAR(32),
			token_expiration CHAR(23)
			)`)
	if err != nil {
		return db, err
	}
	_, err = statement.Exec()
	if err != nil {
		return db, err
	}
	statement, err =
		db.Prepare(`CREATE TABLE IF NOT EXISTS polls (
			poll_id INTEGER PRIMARY KEY,
			email VARCHAR(320), 
			title TEXT
			)`)
	if err != nil {
		return db, err
	}
	_, err = statement.Exec()
	if err != nil {
		return db, err
	}
	statement, err =
		db.Prepare(`CREATE TABLE IF NOT EXISTS questions (
			q_id INTEGER PRIMARY KEY,
			poll_id INTEGER,
			question TEXT
			)`)
	if err != nil {
		return db, err
	}
	_, err = statement.Exec()
	if err != nil {
		return db, err
	}
	statement, err =
		db.Prepare(`CREATE TABLE IF NOT EXISTS answers (
			answer_id INTEGER PRIMARY KEY,
			q_id INTEGER,
			answerInt INTEGER
			)`)
	if err != nil {
		return db, err
	}
	_, err = statement.Exec()
	return db, err
}
