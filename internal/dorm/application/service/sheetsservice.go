package service

import (
	"dorm/internal/common/utils"
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure"
)

type SheetsService struct {
	sheets *infrastructure.Sheets
}

func NewSheetsService(sheets *infrastructure.Sheets) *SheetsService {
	return &SheetsService{
		sheets: sheets,
	}
}

func (s *SheetsService) CreateWeeklySheet(title string, colorInt int, tasks []model.Task) error {
	sheetsColor := utils.IntToSheetsColor(colorInt)
	err := s.sheets.CreateSheet(title, sheetsColor)
	if err != nil {
		return err
	}

	var dataToWrite [][]interface{}

	header := []interface{}{"Задача", "Описание", "Статус"}
	dataToWrite = append(dataToWrite, header)

	for _, task := range tasks {
		row := []interface{}{
			task.AreaID,
			task.Title, // TODO: сделать нормальные данные
		}
		dataToWrite = append(dataToWrite, row)
	}

	return s.sheets.WriteRange(title, "A1", dataToWrite)
}
