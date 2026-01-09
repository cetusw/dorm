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

type SelectAreaState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []dto.TaskViewModel
}

func NewSelectAreaState(u ports.UserUseCase, c ports.CleaningUseCase, t []dto.TaskViewModel) *SelectAreaState {
	return &SelectAreaState{userUseCase: u, cleaningUseCase: c, tasks: t}
}
func (s *SelectAreaState) Name() string { return "SelectArea" }

func (s *SelectAreaState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *SelectAreaState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	if cb.Data == cbBack {
		r.Display(msgDutyManagement+msgDutyManagementDescription+msgSelectAction, taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}
	if strings.HasPrefix(cb.Data, "area:") {
		return s.handleAreaSelection(ctx, cb, r)
	}
	return nil, nil
}

func (s *SelectAreaState) handleAreaSelection(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	r *Responder,
) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	progressText := getCoveredProgress(ctx, s.cleaningUseCase, u.ID())
	areaID, _ := strconv.Atoi(strings.TrimPrefix(cb.Data, "area:"))

	filtered := filterTasksByArea(s.tasks, areaID)
	var areaName string
	if len(filtered) > 0 {
		areaName = filtered[0].AreaName
	} else {
		for _, task := range s.tasks {
			if task.AreaID == areaID {
				areaName = task.AreaName
				break
			}
		}
	}

	if len(filtered) == 0 {
		r.Display(
			fmt.Sprintf("*Взять задачу*\n%s\n%s", progressText, msgNoTasksInArea),
			areaSelectKeyboard(s.tasks),
		)
		return s, nil
	}
	r.Display(
		fmt.Sprintf("*%s*\n%s\n%s", areaName, progressText, msgSelectTask),
		taskSelectKeyboard(filtered),
	)
	return NewSelectTaskState(s.userUseCase, s.cleaningUseCase, s.tasks, areaID), nil
}
