package main

import (
	"fmt"
	"time"
)

// NextDate вычисляет следующую дату для выполнения задачи по правилу повторения.
func NextDate(now time.Time, date string, repeat string) (string, error) {
	const layout = "20060102"

	startDate, err := time.Parse(layout, date)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %w", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	if repeat[0] == 'd' {
		var days int
		_, err := fmt.Sscanf(repeat, "d %d", &days)
		if err != nil {
			return "", fmt.Errorf("некорректное правило повторения: %w", err)
		}
		if days <= 0 || days > 400 {
			return "", fmt.Errorf("недопустимое количество дней: %d", days)
		}
		startDate = startDate.AddDate(0, 0, days)
		for startDate.Before(now) {
			startDate = startDate.AddDate(0, 0, days)
		}
		return startDate.Format(layout), nil
	}

	if repeat == "y" {
		startDate = startDate.AddDate(1, 0, 0)
		for startDate.Before(now) {
			startDate = startDate.AddDate(1, 0, 0)
		}
		return startDate.Format(layout), nil
	}

	return "", fmt.Errorf("неподдерживаемый формат repeat")
}
