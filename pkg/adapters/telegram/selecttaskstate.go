package telegram

import (
	"context"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type SelectTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	allTasks        []dto.TaskViewModel
	areaID          int
}

func NewSelectTaskState(u ports.UserUseCase, c ports.CleaningUseCase, allTasks []dto.TaskViewModel, areaID int) *SelectTaskState {
	return &SelectTaskState{userUseCase: u, cleaningUseCase: c, allTasks: allTasks, areaID: areaID}
}
func (s *SelectTaskState) Name() string { return "SelectTask" }

func (s *SelectTaskState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *SelectTaskState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	if cb.Data == cbBack {
		return s.handleBack(ctx, u, r)
	}
	if strings.HasPrefix(cb.Data, "assign:") {
		return s.handleAssignment(ctx, cb, r, u)
	}
	return nil, nil
}

func (s *SelectTaskState) handleBack(ctx context.Context, user *user.User, r *Responder) (State, error) {
	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())
	r.Display(
		fmt.Sprintf("*Взять задачу*\n%s\n%s", progressText, msgSelectArea),
		areaSelectKeyboard(s.allTasks),
	)
	return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, s.allTasks), nil
}

func (s *SelectTaskState) handleAssignment(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder, user *user.User) (State, error) {
	taskID, _ := uuid.Parse(strings.TrimPrefix(cb.Data, "assign:"))

	var taskToToggle *dto.TaskViewModel
	for _, t := range s.allTasks {
		if t.ID == taskID {
			taskToToggle = &t
			break
		}
	}

	if taskToToggle == nil {
		return s, nil
	}

	var err error
	var notification string
	if taskToToggle.IsAssignedToUser {
		err = s.cleaningUseCase.UnassignTask(ctx, taskID, user.ID())
		notification = msgTaskReturned
	} else {
		err = s.cleaningUseCase.AssignTask(ctx, taskID, user.ID())
		notification = msgTaskAssigned
	}

	if err != nil {
		r.Display(msgAssignTaskErrPrefix+err.Error(), taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	allTasks, _ := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())
	s.allTasks = allTasks

	tasksInSameArea := filterTasksByArea(s.allTasks, s.areaID)
	var areaName string
	if len(tasksInSameArea) > 0 {
		areaName = tasksInSameArea[0].AreaName
	}

	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())
	msg := fmt.Sprintf("*%s*\n%s\n%s %s", areaName, progressText, notification, msgSelectTask)
	r.Display(msg, taskSelectKeyboard(tasksInSameArea))
	return s, nil
}
