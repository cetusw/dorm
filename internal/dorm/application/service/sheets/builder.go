package sheets

import (
	"dorm/internal/common/consts"
	"dorm/internal/dorm/application/model"
	"fmt"
)

type SheetLayout struct {
	TaskTableData       [][]interface{}
	UserTableData       [][]interface{}
	TaskTableRange      string
	UserTableRange      string
	ZoneMergeRanges     map[string][2]int
	TaskCount           int
	UserCount           int
	TaskTableClearRange string
	TaskTableWriteRange string
}

type DutySheetBuilder struct{}

func NewDutySheetBuilder() *DutySheetBuilder {
	return &DutySheetBuilder{}
}

func (b *DutySheetBuilder) Build(order int, tasks []model.DutyTaskView, users []model.User) *SheetLayout {
	taskData, zoneRanges := b.prepareTaskTableData(order, tasks)
	userData := b.prepareUserTableData(users)

	return &SheetLayout{
		TaskTableData:   taskData,
		UserTableData:   userData,
		TaskTableRange:  "A1",
		UserTableRange:  "G2",
		ZoneMergeRanges: zoneRanges,
		TaskCount:       len(tasks),
		UserCount:       len(users) + 1,
	}
}

func (b *DutySheetBuilder) BuildForUpdate(tasks []model.DutyTaskView) *SheetLayout {
	var dataToWrite [][]interface{}
	for _, task := range tasks {
		row := b.formatTaskRow(task)
		dataToWrite = append(dataToWrite, row)
	}

	return &SheetLayout{
		TaskTableData:       dataToWrite,
		TaskTableClearRange: fmt.Sprintf("A%d:K", consts.TasksStartRow),
		TaskTableWriteRange: fmt.Sprintf("A%d", consts.TasksStartRow),
	}
}

func (b *DutySheetBuilder) prepareTaskTableData(order int, tasks []model.DutyTaskView) ([][]interface{}, map[string][2]int) {
	var dataToWrite [][]interface{}
	dataToWrite = append(dataToWrite, []interface{}{fmt.Sprintf(consts.SheetHeaderTeam, order)})
	dataToWrite = append(dataToWrite, []interface{}{
		consts.SheetColumnArea,
		consts.SheetColumnTask,
		consts.SheetColumnCost,
		consts.SheetColumnAssignee,
		consts.SheetColumnState,
	})

	zoneMergeRanges := make(map[string][2]int)

	for i, task := range tasks {
		row := b.formatTaskRow(task)
		dataToWrite = append(dataToWrite, row)

		areaFullName := fmt.Sprintf(consts.SheetRowArea, task.AreaFloor, task.AreaName)
		if ranges, ok := zoneMergeRanges[areaFullName]; ok {
			ranges[1] = consts.TasksStartRow + i
			zoneMergeRanges[areaFullName] = ranges
		} else {
			zoneMergeRanges[areaFullName] = [2]int{consts.TasksStartRow + i, consts.TasksStartRow + i}
		}
	}
	return dataToWrite, zoneMergeRanges
}

func (b *DutySheetBuilder) formatTaskRow(task model.DutyTaskView) []interface{} {
	assignee := consts.DefaultAssignee
	if task.AssigneeFirstName != nil && task.AssigneeLastName != nil {
		assignee = fmt.Sprintf("%s %s.", *task.AssigneeFirstName, string([]rune(*task.AssigneeLastName)[0]))
	}

	state := consts.StateNotDone
	if task.CompletionDate != nil {
		state = consts.StateDone
	}
	if task.VerificationDate != nil {
		state = consts.StateVerified
	}

	return []interface{}{
		fmt.Sprintf("%d этаж. %s", task.AreaFloor, task.AreaName),
		task.TaskTitle,
		task.TaskCost,
		assignee,
		state,
	}
}

func (b *DutySheetBuilder) prepareUserTableData(users []model.User) [][]interface{} {
	var data [][]interface{}
	data = append(data, []interface{}{"Состав команды", "Бронь", "Заработано"})
	for i, user := range users {
		name := fmt.Sprintf("%s %s.", user.FirstName, string([]rune(user.LastName)[0]))
		data = append(data, []interface{}{
			name,
			fmt.Sprintf("=SUMPRODUCT(($D$3:$D = G%d)*$C$3:$C)", i+3),
			fmt.Sprintf("=SUMPRODUCT(($D$3:$D = G%d)*$C$3:$C*($E$3:$E = \"Проверено\"))", i+3),
		})
	}
	data = append(data, []interface{}{
		"Всего на члена команды:",
		fmt.Sprintf("=ROUND(SUM(C:C) / %d, 0)", len(users))})
	return data
}
