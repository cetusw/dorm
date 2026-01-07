package telegram

import (
	"context"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type UnassignSelectTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	allTasks        []ports.TaskViewModel
	areaID          int
}

func NewUnassignSelectTaskState(u ports.UserUseCase, c ports.CleaningUseCase, allTasks []ports.TaskViewModel, areaID int) *UnassignSelectTaskState {
	return &UnassignSelectTaskState{userUseCase: u, cleaningUseCase: c, allTasks: allTasks, areaID: areaID}
}
func (s *UnassignSelectTaskState) Name() string { return "UnassignSelectTask" }

func (s *UnassignSelectTaskState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *UnassignSelectTaskState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	if cb.Data == cbBack {
		return s.handleBack(ctx, u, r)
	}
	if strings.HasPrefix(cb.Data, "unassign:") {
		return s.handleUnassignment(ctx, cb, r, u)
	}
	return nil, nil
}

func (s *UnassignSelectTaskState) handleBack(ctx context.Context, user *user.User, r *Responder) (State, error) {
	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())
	r.Display(
		fmt.Sprintf("*Отдать задачу*\n%s\n%s", progressText, msgSelectArea),
		unassignAreaSelectKeyboard(s.allTasks),
	)
	return NewUnassignSelectAreaState(s.userUseCase, s.cleaningUseCase, s.allTasks), nil
}

func (s *UnassignSelectTaskState) handleUnassignment(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder, user *user.User) (State, error) {
	taskID, _ := uuid.Parse(strings.TrimPrefix(cb.Data, "unassign:"))
	if err := s.cleaningUseCase.UnassignTask(ctx, taskID, user.ID()); err != nil {
		r.Display(msgUnassignTaskErrPrefix+err.Error(), taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())

	allTasks, _ := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
	s.allTasks = allTasks

	tasksInSameArea := filterTasksByArea(s.allTasks, s.areaID)

	if len(tasksInSameArea) > 0 {
		msg := fmt.Sprintf("*%s*\n%s\n%s%s", tasksInSameArea[0].AreaName, progressText, msgTaskUnassigned, msgSelectNextTaskToUnassign)
		r.Display(msg, unassignTaskSelectKeyboard(tasksInSameArea))
		return s, nil
	}
	if len(allTasks) > 0 {
		r.Display(
			fmt.Sprintf("*Отдать задачу*\n%s\n%s", progressText, msgNoTasksInArea),
			unassignAreaSelectKeyboard(allTasks),
		)

		return NewUnassignSelectAreaState(s.userUseCase, s.cleaningUseCase, allTasks), nil
	}
	r.Display(
		fmt.Sprintf("%s%s", msgDutyManagement, msgAllTasksUnassigned),
		taskMenuKeyboard(),
	)
	return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
}
