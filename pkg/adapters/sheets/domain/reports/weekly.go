package reports

import (
	"dorm/pkg/adapters/sheets/domain/widgets"
	"dorm/pkg/adapters/sheets/types"
	"dorm/pkg/core/ports/dto"
)

type WeeklyReportBlueprint struct {
	duty dto.DutyViewModel
}

func NewWeeklyReportBlueprint(duty dto.DutyViewModel) *WeeklyReportBlueprint {
	return &WeeklyReportBlueprint{duty: duty}
}

func (b *WeeklyReportBlueprint) GetTitle() string { return b.duty.DutyName }

func (b *WeeklyReportBlueprint) GetSpreadsheetID() string { return b.duty.SpreadsheetID }

func (b *WeeklyReportBlueprint) GetWidgets() []types.Widget {
	return []types.Widget{
		widgets.NewTaskListWidget(b.duty),
		widgets.NewUserStatsWidget(b.duty),
	}
}
