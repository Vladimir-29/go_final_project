package main

// Task представляет задачу
type Task struct {
	ID      string `json:"id"`      // Уникальный идентификатор задачи
	Date    string `json:"date"`    // Дата задачи в формате YYYYMMDD
	Title   string `json:"title"`   // Заголовок задачи
	Comment string `json:"comment"` // Дополнительный комментарий к задаче
	Repeat  string `json:"repeat"`  // Условие повторения задачи
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}
