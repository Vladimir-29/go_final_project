package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/vladimir-29/go_final_project/models"

	_ "modernc.org/sqlite" //
)

type Database struct {
	Conn *sql.DB
}

// NewDatabase создает новое подключение к базе данных
func NewDatabase(dataSourceName string) (*Database, error) {
	conn, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Проверка соединения с базой данных
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &Database{Conn: conn}, nil
}

// Close закрывает соединение с базой данных
func Close(db *Database) error {
	return db.Conn.Close()
}

func GetTaskByID(db *sql.DB, id int64) ([]byte, int, error) {
	var task models.Task
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"

	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.StatusNotFound, fmt.Errorf("задача не найдена")
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	response, err := json.Marshal(task)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	return response, http.StatusOK, nil
}

func GetTasks(db *sql.DB, limit int) ([]byte, int, error) {
	var tasks []models.Task

	query := "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?"
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		tasks = append(tasks, task)
	}

	err = rows.Err()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	response, err := json.Marshal(map[string][]models.Task{"tasks": tasks})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}
