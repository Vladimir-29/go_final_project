package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite" // SQLite3 driver
)

// Функция для получения задачи из базы данных по идентификатору
func GetTaskByID(db *sql.DB, id string) ([]byte, int, error) {
	var task Task
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"

	var taskID int64
	err := db.QueryRow(query, id).Scan(&taskID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.StatusNotFound, fmt.Errorf("задача не найдена")
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}

	task.ID = fmt.Sprintf("%d", taskID)

	response, err := json.Marshal(task)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	return response, http.StatusOK, nil
}
