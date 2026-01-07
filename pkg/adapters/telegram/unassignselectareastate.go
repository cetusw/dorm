package telegram

import (
	"context"
	"dorm/pkg/core/ports"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UnassignSelectAreaState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []ports.TaskViewModel
}

func NewUnassignSelectAreaState(u ports.UserUseCase, c ports.CleaningUseCase, t []ports.TaskViewModel) *UnassignSelectAreaState {
	return &UnassignSelectAreaState{userUseCase: u, cleaningUseCase: c, tasks: t}
}
func (s *UnassignSelectAreaState) Name() string { return "UnassignSelectArea" }

func (s *UnassignSelectAreaState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *UnassignSelectAreaState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	if cb.Data == cbBack {
		r.Display(
			fmt.Sprintf("%s%s%s", msgDutyManagement, msgDutyManagementDescription, msgSelectAction),
			taskMenuKeyboard(),
		)
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}
	if strings.HasPrefix(cb.Data, "unassign_area:") {
		return s.handleAreaSelection(ctx, cb, r)
	}
	return nil, nil
}

func (s *UnassignSelectAreaState) handleAreaSelection(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	progressText := getCoveredProgress(ctx, s.cleaningUseCase, u.ID())
	areaID, _ := strconv.Atoi(strings.TrimPrefix(cb.Data, "unassign_area:"))

	filtered := filterTasksByArea(s.tasks, areaID)

	if len(filtered) == 0 {
		r.Display(
			fmt.Sprintf("*Отдать задачу*\n%s\n%s", progressText, msgNoTasksInArea),
			unassignAreaSelectKeyboard(s.tasks),
		)
		return s, nil
	}
	r.Display(
		fmt.Sprintf("*%s*\n%s\n%s", filtered[0].AreaName, progressText, msgSelectTask),
		unassignTaskSelectKeyboard(filtered),
	)
	return NewUnassignSelectTaskState(s.userUseCase, s.cleaningUseCase, s.tasks, areaID), nil
}
