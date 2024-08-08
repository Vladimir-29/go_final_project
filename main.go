package main

import (
	"log"
	"net/http"
	"path/filepath"
)

func main() {

	// Инициализация базы данных
	if err := InitializeDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}

	// Директория для статических файлов веб-приложения
	webDir := "./web"

	// Обработчик для статических файлов
	fileHandler := http.FileServer(http.Dir(webDir))

	// Настройка маршрутов для статических файлов
	http.Handle("/js/", fileHandler)
	http.Handle("/css/", fileHandler)
	http.Handle("/favicon.ico", fileHandler)

	// Обработчик для страницы login.html
	http.HandleFunc("/login.html", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(webDir, "login.html"))
	})

	// Обработчик для корневого пути и index.html
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" || req.URL.Path == "/index.html" {
			http.ServeFile(w, req, filepath.Join(webDir, "index.html"))
		} else {
			http.NotFound(w, req)
		}
	})

	// Обработчик для API запросов
	http.HandleFunc("/api/tasks", GetTasksHandler)     // Получение списка задач
	http.HandleFunc("/api/task/done", TaskDoneHandler) // Завершение задачи
	http.HandleFunc("/api/nextdate", NextDateHandler)  // Получение следующей даты
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetTaskByIDHandler(w, r) // Получение задачи по ID
		case http.MethodPost:
			AddTaskHandler(w, r) // Добавление новой задачи
		case http.MethodPut:
			UpdateTaskHandler(w, r) // Обновление задачи
		case http.MethodDelete:
			DeleteTaskHandler(w, r) // Удаление задачи
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Запуск HTTP сервера
	err := http.ListenAndServe("localhost:7540", nil)
	if err != nil {
		panic(err)
	}
}
