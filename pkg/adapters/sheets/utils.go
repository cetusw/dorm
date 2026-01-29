package sheets

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"google.golang.org/api/sheets/v4"

	"dorm/pkg/core/ports/dto"
)

type RowRange struct {
	Start int64
	End   int64
}

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

func DarkenColor(c *sheets.Color) *sheets.Color {
	factor := 0.7
	return &sheets.Color{
		Red: c.Red * factor, Green: c.Green * factor, Blue: c.Blue * factor,
	}
}

func FormatMemberName(userStats *dto.UserStats, usersStats []*dto.UserStats) string {
	if userStats == nil {
		return "Никто"
	}
	initial, _ := utf8.DecodeRuneInString(userStats.LastName)
	if initial == utf8.RuneError {
		return userStats.FirstName
	}

	shortName := fmt.Sprintf("%s %c.", userStats.FirstName, initial)

	count := 0
	for _, m := range usersStats {
		mInitial, _ := utf8.DecodeRuneInString(m.LastName)
		mShortName := fmt.Sprintf("%s %c.", m.FirstName, mInitial)

		if mShortName == shortName {
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

func SortForTableView(tasks []dto.TaskViewModel) []dto.TaskViewModel {
	sortedTasks := slices.Clone(tasks)
	slices.SortFunc(sortedTasks, func(a, b dto.TaskViewModel) int {
		if a.AreaFloor != b.AreaFloor {
			return b.AreaFloor - a.AreaFloor
		}
		if a.AreaName != b.AreaName {
			return strings.Compare(a.AreaName, b.AreaName)
		}
		return b.Cost - a.Cost
	})
	return sortedTasks
}
