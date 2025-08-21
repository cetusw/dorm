package service

import (
	"dorm/internal/common/consts"
	"dorm/internal/common/utils"
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure"
	"fmt"
)

type SheetsService struct {
	sheets *infrastructure.Sheets
}

func NewSheetsService(sheets *infrastructure.Sheets) *SheetsService {
	return &SheetsService{
		sheets: sheets,
	}
}

func (s *SheetsService) CreateWeeklySheet(title string, hexColor string, teamID int, dutyTasksReadable []model.DutyTaskReadable) error {
	sheetsColor, err := utils.HexToSheetsColor(hexColor)
	if err != nil {
		return err
	}
	err = s.sheets.CreateSheet(title, sheetsColor)
	if err != nil {
		return err
	}

	var dataToWrite [][]interface{}

	firstRowHeader := []interface{}{fmt.Sprintf(consts.CurrentTeamID, teamID)}
	secondRowHeader := []interface{}{consts.Area, consts.Task, consts.Cost, consts.Assignee, consts.State}
	dataToWrite = append(dataToWrite, firstRowHeader, secondRowHeader)

	for _, dutyTaskReadable := range dutyTasksReadable {
		assignee := "Никто"
		if dutyTaskReadable.AssigneeFirstName != nil && dutyTaskReadable.AssigneeLastName != nil {
			firstName := *dutyTaskReadable.AssigneeFirstName
			lastName := *dutyTaskReadable.AssigneeLastName
			assignee = fmt.Sprintf("%s %s.", firstName, string([]rune(lastName)[0]))
		}

		state := "Не сделано"
		if dutyTaskReadable.CompletionDate != nil {
			state = "Сделано"
		}
		if dutyTaskReadable.VerificationDate != nil {
			state = "Проверено"
		}

		row := []interface{}{
			fmt.Sprintf("%d этаж. %s", dutyTaskReadable.AreaFloor, dutyTaskReadable.AreaName),
			dutyTaskReadable.TaskTitle,
			dutyTaskReadable.TaskCost,
			assignee,
			state,
		}
		dataToWrite = append(dataToWrite, row)
	}

	return s.sheets.WriteRange(title, "A1", dataToWrite)
}
