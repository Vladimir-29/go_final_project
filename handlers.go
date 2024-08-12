package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vladimir-29/go_final_project/database"
	"github.com/vladimir-29/go_final_project/models"

	_ "modernc.org/sqlite" // SQLite3 driver
)

// writeErrorResponse отправляет ошибочный ответ с указанным статусом
func writeErrorResponse(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: errorMsg})
}

// AddTaskHandler обрабатывает POST-запросы для добавления новой задачи
func AddTaskHandler(db *database.Database, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "AddTaskHandler(): Method not supported", http.StatusMethodNotAllowed)
		return
	}

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeErrorResponse(w, "AddTaskHandler(): JSON deserialization error", http.StatusBadRequest)
		return
	}

	// Проверяем обязательное поле title
	if task.Title == "" {
		writeErrorResponse(w, "AddTaskHandler(): Task title not specified", http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	// Проверяем формат даты
	date, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeErrorResponse(w, "AddTaskHandler(): Date is not in the correct format", http.StatusBadRequest)
		return
	}

	// Проверка формата поля Repeat
	if task.Repeat != "" {
		dateCheck, err := NextDate(time.Now(), task.Date, task.Repeat)
		if dateCheck == "" && err != nil {
			writeErrorResponse(w, "AddTaskHandler() Invalid repetition condition", http.StatusBadRequest)
			return
		}
	}

	now := time.Now()
	if date.Before(now) {
		if task.Repeat == "" || date.Truncate(24*time.Hour) == date.Truncate(24*time.Hour) {
			task.Date = time.Now().Format("20060102")
		} else {
			dateStr := date.Format("20060102")
			nextDate, err := NextDate(now, dateStr, task.Repeat)
			if err != nil {
				writeErrorResponse(w, "AddTaskHandler(): Wrong repetition rule", http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	}

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.Conn.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		writeErrorResponse(w, "AddTaskHandler(): Error executing request", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		writeErrorResponse(w, "AddTaskHandler(): Error getting task ID", http.StatusInternalServerError)
		return
	}

	task.ID = fmt.Sprint(id)

	response := map[string]interface{}{"id": id}
	responseId, err := json.Marshal(response)
	if err != nil {
		writeErrorResponse(w, "AddTaskHandler(): JSON encoding error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseId)
}

// NextDateHandler обрабатывает запросы на получение следующей даты по повторению
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	const layout = "20060102"

	now, err := time.Parse(layout, nowStr)
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Некорректная текущая дата: %v", err), http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Ошибка в NextDate: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, nextDate)
}

// GetTasksHandler обрабатывает запросы на получение списка задач
func GetTasksHandler(db *database.Database, w http.ResponseWriter, r *http.Request) {
	tasks, status, err := database.GetTasks(db.Conn, 50)
	if err != nil {
		writeErrorResponse(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(tasks)
}

// GetTaskByIDHandler обрабатывает запросы на получение задачи по ID
func GetTaskByIDHandler(db *database.Database, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeErrorResponse(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Преобразование строки в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorResponse(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	task, status, err := database.GetTaskByID(db.Conn, id)
	if err != nil {
		writeErrorResponse(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(task)
}

// UpdateTaskHandler обрабатывает PUT-запросы на обновление задачи
func UpdateTaskHandler(db *database.Database, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeErrorResponse(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeErrorResponse(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		writeErrorResponse(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeErrorResponse(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if task.Date != "" {
		if _, err := time.Parse("20060102", task.Date); err != nil {
			writeErrorResponse(w, "Некорректный формат даты", http.StatusBadRequest)
			return
		}
	}

	now := time.Now().Format("20060102")
	if task.Date == "" {
		task.Date = now
	} else {
		taskDate, _ := time.Parse("20060102", task.Date)
		nowDate, _ := time.Parse("20060102", now)
		if taskDate.Before(nowDate) {
			if task.Repeat == "" {
				task.Date = now
			} else {
				nextDate, err := NextDate(nowDate, task.Date, task.Repeat)
				if err != nil {
					writeErrorResponse(w, fmt.Sprintf("Ошибка вычисления следующей даты: %v", err), http.StatusInternalServerError)
					return
				}
				task.Date = nextDate
			}
		}
	}

	if task.Repeat != "" {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeErrorResponse(w, "Неправильный формат повторения", http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	}

	// Обновление задачи
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	result, err := db.Conn.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Ошибка обновления задачи: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Ошибка проверки количества затронутых строк: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		writeErrorResponse(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

// TaskDoneHandler обрабатывает запросы для завершения задачи
func TaskDoneHandler(db *database.Database, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "Method not supported", http.StatusBadRequest)
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeErrorResponse(w, "Task ID not specified", http.StatusBadRequest)
		return
	}

	idParamNum, err := strconv.Atoi(idParam)
	if err != nil {
		writeErrorResponse(w, "Invalid task ID format", http.StatusBadRequest)
		return
	}

	var task models.Task
	var id int64
	err = db.Conn.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", idParamNum).Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if errors.Is(err, sql.ErrNoRows) {
		writeErrorResponse(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		writeErrorResponse(w, "Error retrieving task data", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	if task.Repeat != "" {
		newTaskDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeErrorResponse(w, "Incorrect task repetition condition", http.StatusBadRequest)
			return
		}
		task.Date = newTaskDate

		_, err = db.Conn.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, id)
		if err != nil {
			writeErrorResponse(w, "Error updating task", http.StatusInternalServerError)
			return
		}
	} else {
		result, err := db.Conn.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			writeErrorResponse(w, "Error deleting task", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			writeErrorResponse(w, "Unable to determine the number of rows affected after deleting a task", http.StatusInternalServerError)
			return
		} else if rowsAffected == 0 {
			writeErrorResponse(w, "Task not found", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

// DeleteTaskHandler обрабатывает DELETE-запросы для удаления задачи
func DeleteTaskHandler(db *database.Database, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrorResponse(w, "DeleteTaskHandler(): method not supported", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.FormValue("id")
	if idParam == "" {
		writeErrorResponse(w, "DeleteTaskHandler(): Task ID not specified", http.StatusBadRequest)
		return
	}

	idParamNum, err := strconv.Atoi(idParam)
	if err != nil {
		writeErrorResponse(w, "DeleteTaskHandler(): Error converting idParam to number", http.StatusInternalServerError)
		return
	}

	query := "DELETE FROM scheduler WHERE id = ?"
	result, err := db.Conn.Exec(query, idParamNum)
	if err != nil {
		writeErrorResponse(w, "DeleteTaskHandler(): Error deleting task", http.StatusInternalServerError)
		return
	}

	// Проверка количества затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeErrorResponse(w, "DeleteTaskHandler(): Unable to determine the number of rows affected after deleting a task", http.StatusInternalServerError)
		return
	} else if rowsAffected == 0 {
		writeErrorResponse(w, "DeleteTaskHandler(): Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}
