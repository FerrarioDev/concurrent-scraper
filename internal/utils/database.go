package utils

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var schema = `
    CREATE TABLE IF NOT EXISTS sites (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT NOT NULL,
        title TEXT NOT NULL,
        links INTEGER,
        father_id INTEGER,
        FOREIGN KEY (father_id) REFERENCES sites(id)
    );
`

func InitDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "scraper.sqlite")
	if err != nil {
		return nil
	}

	db.MustExec(schema)

	return db
}
