package sheets

import (
	"dorm/internal/common/consts"
	"strings"
)

type ParsedTaskDTO struct {
	AreaName     string
	TaskTitle    string
	AssigneeName string
	StatusRaw    string
}

type DutySheetParser struct{}

func NewDutySheetParser() *DutySheetParser {
	return &DutySheetParser{}
}

func (p *DutySheetParser) Parse(data [][]interface{}) ([]ParsedTaskDTO, error) {
	var result []ParsedTaskDTO

	const (
		colArea     = 0
		colTask     = 1
		colAssignee = 3
		colStatus   = 4
	)

	currentArea := ""

	startRowIndex := consts.TasksStartRow - 1

	if len(data) <= startRowIndex {
		return nil, nil
	}

	for i := startRowIndex; i < len(data); i++ {
		row := data[i]

		if len(row) <= colTask {
			continue
		}

		areaVal := getString(row, colArea)
		if areaVal != "" {
			currentArea = areaVal
		}

		taskTitle := getString(row, colTask)
		if taskTitle == "" {
			continue
		}

		dto := ParsedTaskDTO{
			AreaName:     currentArea,
			TaskTitle:    taskTitle,
			AssigneeName: getString(row, colAssignee),
			StatusRaw:    getString(row, colStatus),
		}
		result = append(result, dto)
	}

	return result, nil
}

func getString(row []interface{}, index int) string {
	if index >= len(row) {
		return ""
	}
	val, ok := row[index].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(val)
}
