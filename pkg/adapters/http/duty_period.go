package http

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func parseDutyPeriod(startDate, endDate string) (time.Time, time.Time, error) {
	location := residentDutyWeekLocation()

	parsedStart, err := time.ParseInLocation("2006-01-02", startDate, location)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "некорректная дата начала")
	}

	parsedEnd, err := time.ParseInLocation("2006-01-02", endDate, location)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "некорректная дата окончания")
	}

	normalizedStart := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 9, 0, 0, 0, location)
	normalizedEnd := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 9, 0, 0, 0, location)

	if !normalizedStart.Before(normalizedEnd) {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "дата начала должна быть раньше даты окончания")
	}

	return normalizedStart, normalizedEnd, nil
}

func residentDutyWeekLocation() *time.Location {
	timezone := os.Getenv("TZ")
	if timezone == "" {
		return time.Local
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Local
	}

	return location
}
