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

type SelectTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	allTasks        []ports.TaskViewModel
	areaID          int
}

func NewSelectTaskState(u ports.UserUseCase, c ports.CleaningUseCase, allTasks []ports.TaskViewModel, areaID int) *SelectTaskState {
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
	if err := s.cleaningUseCase.AssignTask(ctx, taskID, user.ID()); err != nil {
		r.Display(msgAssignTaskErrPrefix+err.Error(), taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())

	allTasks, _ := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())
	s.allTasks = allTasks

	tasksInSameArea := filterTasksByArea(s.allTasks, s.areaID)
	var areaName string
	if len(s.allTasks) > 0 {
		for _, task := range s.allTasks {
			if task.AreaID == s.areaID {
				areaName = task.AreaName
				break
			}
		}
	}

	if len(tasksInSameArea) > 0 {
		msg := fmt.Sprintf("*%s*\n%s\n%s %s", areaName, progressText, msgTaskAssigned, msgTakeNextTask)
		r.Display(msg, taskSelectKeyboard(tasksInSameArea))
		return s, nil
	}
	if len(s.allTasks) > 0 {
		r.Display(
			fmt.Sprintf("*Взять задачу*\n%s\n%s", progressText, msgNoTasksInArea),
			areaSelectKeyboard(s.allTasks),
		)
		return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, s.allTasks), nil
	}
	r.Display(
		fmt.Sprintf("%s%s", msgDutyManagement, msgAllTasksUnassigned),
		taskMenuKeyboard(),
	)
	return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
}
