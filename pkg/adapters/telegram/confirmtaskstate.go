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

type ConfirmTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	allTasks        []ports.TaskViewModel
	areaID          int
}

func NewConfirmTaskState(u ports.UserUseCase, c ports.CleaningUseCase, allTasks []ports.TaskViewModel, areaID int) *ConfirmTaskState {
	return &ConfirmTaskState{userUseCase: u, cleaningUseCase: c, allTasks: allTasks, areaID: areaID}
}
func (s *ConfirmTaskState) Name() string { return "ConfirmTask" }

func (s *ConfirmTaskState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *ConfirmTaskState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	if cb.Data == cbBack {
		return s.handleBack(ctx, u, r)
	}
	if strings.HasPrefix(cb.Data, "complete:") {
		return s.handleCompletion(ctx, cb, r, u)
	}
	return nil, nil
}

func (s *ConfirmTaskState) handleBack(ctx context.Context, user *user.User, r *Responder) (State, error) {
	progressText := getUserProgress(ctx, s.cleaningUseCase, user.ID())
	r.Display(progressText+"\n"+msgSelectConfirmArea, confirmAreaSelectKeyboard(s.allTasks))
	return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, s.allTasks), nil
}

func (s *ConfirmTaskState) handleCompletion(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder, user *user.User) (State, error) {
	taskID, _ := uuid.Parse(strings.TrimPrefix(cb.Data, "complete:"))

	var taskToToggle *ports.TaskViewModel
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
	if taskToToggle.IsDone {
		err = s.cleaningUseCase.OpenTask(ctx, taskID, user.ID())
		notification = msgTaskUncompleted
	} else {
		err = s.cleaningUseCase.CompleteTask(ctx, taskID, user.ID())
		notification = msgTaskCompleted
	}

	if err != nil {
		r.Display(msgCompleteTaskErrPrefix+err.Error(), taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	allAssigned, _ := s.cleaningUseCase.GetAllAssignedTasks(ctx, user.ID())
	s.allTasks = allAssigned

	tasksInSameArea := filterTasksByArea(s.allTasks, s.areaID)
	var areaName string
	if len(tasksInSameArea) > 0 {
		areaName = tasksInSameArea[0].AreaName
	}

	progressText := getUserProgress(ctx, s.cleaningUseCase, user.ID())
	msg := fmt.Sprintf("*%s*\n%s\n%s %s", areaName, progressText, notification, msgSelectConfirmTask)
	r.Display(msg, taskConfirmKeyboard(tasksInSameArea))
	return s, nil
}
