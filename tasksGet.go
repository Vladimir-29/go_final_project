package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite" // SQLite3 driver
)

// GetTasks возвращает список ближайших задач
func GetTasks(db *sql.DB, limit int) ([]byte, int, error) {
	var tasks []Task

	query := "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?"
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Task
		var id int64
		err := rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		t.ID = fmt.Sprintf("%d", id)
		tasks = append(tasks, t)
	}

	err = rows.Err()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if tasks == nil {
		tasks = []Task{}
	}

	response, err := json.Marshal(map[string][]Task{"tasks": tasks})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}
