package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/vladimir-29/go_final_project/database"

	_ "modernc.org/sqlite" // SQLite3 driver
)

func main() {

	// Инициализация базы данных
	if err := database.InitializeDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}

	db, err := database.NewDatabase("./scheduler.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

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

	// /api/tasks - Получение списка задач (только GET запросы)
	http.HandleFunc("GET /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		GetTasksHandler(db, w, r)
	})

	// /api/task/done - Завершение задачи (только POST запросы)
	http.HandleFunc("POST /api/task/done", func(w http.ResponseWriter, r *http.Request) {
		TaskDoneHandler(db, w, r)
	})
	// /api/task - Обработка различных HTTP методов (GET, POST, PUT, DELETE)
	http.HandleFunc("GET /api/task", func(w http.ResponseWriter, r *http.Request) {
		GetTaskByIDHandler(db, w, r)
	})
	http.HandleFunc("POST /api/task", func(w http.ResponseWriter, r *http.Request) {
		AddTaskHandler(db, w, r)
	})
	http.HandleFunc("PUT /api/task", func(w http.ResponseWriter, r *http.Request) {
		UpdateTaskHandler(db, w, r)
	})
	http.HandleFunc("DELETE /api/task", func(w http.ResponseWriter, r *http.Request) {
		DeleteTaskHandler(db, w, r)
	})

	// /api/nextdate - Получение следующей даты (только GET запросы)
	http.HandleFunc("GET /api/nextdate", NextDateHandler)

	// Запуск HTTP сервера
	err = http.ListenAndServe("localhost:7540", nil)
	if err != nil {
		panic(err)
	}
}
