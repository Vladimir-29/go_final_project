package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitializeDatabase инициализирует базу данных и создает таблицу, если её нет
func InitializeDatabase() error {
	appPath, err := os.Executable()
	if err != nil {
		return err
	}

	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	install := false
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	if install {
		createTableSQL := `CREATE TABLE scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat VARCHAR(128) NOT NULL
	);`

		_, err = db.Exec(createTableSQL)
		if err != nil {
			return err
		}

		createIndexSQL := `CREATE INDEX idx_scheduler_date ON scheduler(date);`
		_, err = db.Exec(createIndexSQL)
		if err != nil {
			return err
		}
	}

	return nil
}
