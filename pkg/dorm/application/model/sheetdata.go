package model

type SheetData struct {
	SpreadsheetID string
	Title         string
	TeamColor     string
	Order         int
	Tasks         []DutyTaskView
	Users         []User
}
