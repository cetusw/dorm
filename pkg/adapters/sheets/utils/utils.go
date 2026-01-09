package utils

import (
	"dorm/pkg/core/ports/dto"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

func HexToRGB(hexStr string) (*sheets.Color, error) {
	var r, g, b uint8
	_, err := fmt.Sscanf(hexStr, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return &sheets.Color{Red: 1, Green: 1, Blue: 1}, err
	}
	return &sheets.Color{
		Red:   float64(r) / 255,
		Green: float64(g) / 255,
		Blue:  float64(b) / 255,
	}, nil
}

func FormatMemberName(userStats *dto.UserStats, usersStats []*dto.UserStats) string {
	if userStats == nil {
		return "Никто"
	}
	shortName := fmt.Sprintf("%s %s.", userStats.FirstName, string([]rune(userStats.LastName)[0]))

	count := 0
	for _, m := range usersStats {
		name := fmt.Sprintf("%s %s.", m.FirstName, string([]rune(m.LastName)[0]))
		if name == shortName {
			count++
		}
	}

	if count > 1 {
		return fmt.Sprintf("%s %s", userStats.FirstName, userStats.LastName)
	}
	return shortName
}

func FormatStatus(task dto.TaskViewModel) string {
	if task.IsVerified {
		return "Проверено"
	}
	if task.IsCompleted {
		return "Сделано"
	}
	return "Не сделано"
}
