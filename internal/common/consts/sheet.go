package consts

import "google.golang.org/api/sheets/v4"

const (
	SheetHeaderTeam = "На этой неделе убирается команда %d"
)

const (
	SheetColumnArea     = "Зона"
	SheetColumnTask     = "Задача"
	SheetColumnCost     = "Стоимость"
	SheetColumnAssignee = "Исполнитель"
	SheetColumnState    = "Статус"
)

const (
	SheetRowArea = "%d этаж. %s"
)

const (
	DefaultAssignee = "Никто"
)

const (
	StateNotDone  = "Не сделано"
	StateDone     = "Сделано"
	StateVerified = "Проверено"
)

const (
	TasksStartRow = 3
)

var (
	SubHeaderBgColor = &sheets.Color{Red: 0.9, Green: 0.9, Blue: 0.9}

	CostMinColor = &sheets.Color{Red: 0.76, Green: 0.92, Blue: 0.76}
	CostMidColor = &sheets.Color{Red: 1.0, Green: 0.92, Blue: 0.6}
	CostMaxColor = &sheets.Color{Red: 0.96, Green: 0.76, Blue: 0.76}
)
