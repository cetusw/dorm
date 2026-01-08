package telegram

import (
	"context"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ConfirmSelectAreaState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []dto.TaskViewModel
}

func NewConfirmSelectAreaState(u ports.UserUseCase, c ports.CleaningUseCase, t []dto.TaskViewModel) *ConfirmSelectAreaState {
	return &ConfirmSelectAreaState{userUseCase: u, cleaningUseCase: c, tasks: t}
}
func (s *ConfirmSelectAreaState) Name() string { return "ConfirmSelectArea" }

func (s *ConfirmSelectAreaState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *ConfirmSelectAreaState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	if cb.Data == cbBack {
		r.Display(msgDutyManagement+msgDutyManagementDescription+msgSelectAction, taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}
	if strings.HasPrefix(cb.Data, "conf_area:") {
		return s.handleAreaSelection(ctx, cb, r)
	}
	return nil, nil
}

func (s *ConfirmSelectAreaState) handleAreaSelection(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	progressText := getUserProgress(ctx, s.cleaningUseCase, u.ID())
	areaID, _ := strconv.Atoi(strings.TrimPrefix(cb.Data, "conf_area:"))
	filtered := filterTasksByArea(s.tasks, areaID)

	if len(filtered) == 0 {
		r.Display(
			fmt.Sprintf("*Подтвердить выполнение*\n%s\n%s", progressText, msgNoTasksInArea),
			confirmAreaSelectKeyboard(s.tasks),
		)
		return s, nil
	}

	r.Display(fmt.Sprintf("*%s*\n%s\n%s", filtered[0].AreaName, progressText, msgSelectConfirmTask), taskConfirmKeyboard(filtered))
	return NewConfirmTaskState(s.userUseCase, s.cleaningUseCase, s.tasks, areaID), nil
}
