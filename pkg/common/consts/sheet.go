package consts

import "google.golang.org/api/sheets/v4"

const (
	DefaultFontFamily       = "Montserrat"
	DefaultFontSize         = 10
	TaskTableHeaderFontSize = 14
	TaskTableStartRow       = 0
	TaskTableHeaderHeight   = 60
	TasksStartRow           = 3
	TasksTableRange         = "A1:E"
	UserTableRange          = "G2:I"
	ClearRange              = "A1:K"
)

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

var (
	SubHeaderBgColor     = &sheets.Color{Red: 0.9, Green: 0.9, Blue: 0.9}
	TasksBackgroundColor = &sheets.Color{Red: 0.953, Green: 0.953, Blue: 0.953}
	CostMinColor         = &sheets.Color{Red: 0.76, Green: 0.92, Blue: 0.76}
	CostMidColor         = &sheets.Color{Red: 1.0, Green: 0.92, Blue: 0.6}
	CostMaxColor         = &sheets.Color{Red: 0.96, Green: 0.76, Blue: 0.76}
)
