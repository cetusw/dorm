package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports"
)

type TaskMenuState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
}

func NewTaskMenuState(u ports.UserUseCase, c ports.CleaningUseCase) *TaskMenuState {
	return &TaskMenuState{userUseCase: u, cleaningUseCase: c}
}
func (s *TaskMenuState) Name() string { return "TaskMenu" }

func (s *TaskMenuState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *TaskMenuState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, err := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	if err != nil {
		return nil, err
	}
	if onDuty, err := s.cleaningUseCase.IsUserOnDuty(ctx, u.ID()); err != nil || !onDuty {
		r.Display(msgTeamNotOnDutyError, nil)
		return NewMainMenuState(s.userUseCase, s.cleaningUseCase), err
	}

	switch cb.Data {
	case cbAssign:
		return s.handleAssign(ctx, u, r)
	case cbUnassign:
		return s.handleUnassign(ctx, u, r)
	case cbConfirm:
		return s.handleConfirm(ctx, u, r)
	}
	return nil, nil
}

func (s *TaskMenuState) handleAssign(ctx context.Context, user *user.User, r *Responder) (State, error) {
	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())
	tasks, err := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		r.Display(msgDutyManagement+msgNoFreeTasks, taskMenuKeyboard())
		return s, nil
	}
	r.Display(
		fmt.Sprintf("*Взять задачу*\n%s\n%s", progressText, msgSelectArea),
		areaSelectKeyboard(tasks),
	)
	return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
}

func (s *TaskMenuState) handleUnassign(ctx context.Context, user *user.User, r *Responder) (State, error) {
	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())
	tasks, err := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		r.Display(
			fmt.Sprintf("%s%s", msgDutyManagement, msgNoActiveTasks),
			taskMenuKeyboard(),
		)
		return s, nil
	}
	r.Display(
		fmt.Sprintf("*Отдать задачу*\n%s\n%s", progressText, msgSelectArea),
		unassignAreaSelectKeyboard(tasks),
	)
	return NewUnassignSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
}

func (s *TaskMenuState) handleConfirm(ctx context.Context, user *user.User, r *Responder) (State, error) {
	progressText := getUserProgress(ctx, s.cleaningUseCase, user.ID())
	tasks, err := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		r.Display(msgDutyManagement+msgNoActiveTasks, taskMenuKeyboard())
		return s, nil
	}
	r.Display(
		fmt.Sprintf("*Подтвердить выполнение*\n%s\n%s", progressText, msgSelectConfirmArea),
		confirmAreaSelectKeyboard(tasks),
	)
	return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
}
