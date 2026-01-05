package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"

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
	r.Display(progressText+msgSelectArea, areaSelectKeyboard(tasks))
	return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
}

func (s *TaskMenuState) handleConfirm(ctx context.Context, user *user.User, r *Responder) (State, error) {
	progressText := getMetProgress(ctx, s.cleaningUseCase, user.ID())
	tasks, err := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		r.Display(msgDutyManagement+msgNoActiveTasks, taskMenuKeyboard())
		return s, nil
	}
	r.Display(progressText+msgSelectConfirmArea, confirmAreaSelectKeyboard(tasks))
	return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
}

type SelectAreaState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []ports.TaskViewModel
}

func NewSelectAreaState(u ports.UserUseCase, c ports.CleaningUseCase, t []ports.TaskViewModel) *SelectAreaState {
	return &SelectAreaState{userUseCase: u, cleaningUseCase: c, tasks: t}
}
func (s *SelectAreaState) Name() string { return "SelectArea" }

func (s *SelectAreaState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *SelectAreaState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	if cb.Data == cbBack {
		r.Display(msgDutyManagement+msgSelectAction, taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}
	if strings.HasPrefix(cb.Data, "area:") {
		return s.handleAreaSelection(ctx, cb, r)
	}
	return nil, nil
}

func (s *SelectAreaState) handleAreaSelection(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	progressText := getCoveredProgress(ctx, s.cleaningUseCase, u.ID())
	areaID, _ := strconv.Atoi(strings.TrimPrefix(cb.Data, "area:"))

	var filtered []ports.TaskViewModel
	for _, t := range s.tasks {
		if t.AreaID == areaID {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		r.Display(progressText+msgNoTasksInArea, areaSelectKeyboard(s.tasks))
		return s, nil
	}
	r.Display(fmt.Sprintf("*%s*\n%s%s", filtered[0].AreaName, progressText, msgSelectTask), taskSelectKeyboard(filtered))
	return NewSelectTaskState(s.userUseCase, s.cleaningUseCase, filtered), nil
}

type SelectTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []ports.TaskViewModel
}

func NewSelectTaskState(u ports.UserUseCase, c ports.CleaningUseCase, tasks []ports.TaskViewModel) *SelectTaskState {
	return &SelectTaskState{userUseCase: u, cleaningUseCase: c, tasks: tasks}
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
	tasks, err := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())
	if err != nil {
		return nil, err
	}
	r.Display(progressText+msgSelectArea, areaSelectKeyboard(tasks))
	return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
}

func (s *SelectTaskState) handleAssignment(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder, user *user.User) (State, error) {
	taskID, _ := uuid.Parse(strings.TrimPrefix(cb.Data, "assign:"))
	if err := s.cleaningUseCase.AssignTask(ctx, taskID, user.ID()); err != nil {
		r.Display(msgAssignTaskErrPrefix+err.Error(), taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	progressText := getCoveredProgress(ctx, s.cleaningUseCase, user.ID())
	areaID := s.tasks[0].AreaID
	allTasks, _ := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())
	var tasksInSameArea []ports.TaskViewModel
	for _, t := range allTasks {
		if t.AreaID == areaID {
			tasksInSameArea = append(tasksInSameArea, t)
		}
	}

	if len(tasksInSameArea) > 0 {
		msg := fmt.Sprintf("*%s*\n%s%s%s", tasksInSameArea[0].AreaName, progressText, msgTaskAssigned, msgTakeNextTask)
		r.Display(msg, taskSelectKeyboard(tasksInSameArea))
		return s, nil
	}
	if len(allTasks) > 0 {
		msg := progressText + msgTaskAssigned + msgNoMoreTasksInArea
		r.Display(msg, areaSelectKeyboard(allTasks))
		return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, allTasks), nil
	}
	r.Display(progressText+msgNoMoreFreeTasks, taskMenuKeyboard())
	return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
}

type ConfirmSelectAreaState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []ports.TaskViewModel
}

func NewConfirmSelectAreaState(u ports.UserUseCase, c ports.CleaningUseCase, t []ports.TaskViewModel) *ConfirmSelectAreaState {
	return &ConfirmSelectAreaState{userUseCase: u, cleaningUseCase: c, tasks: t}
}
func (s *ConfirmSelectAreaState) Name() string { return "ConfirmSelectArea" }

func (s *ConfirmSelectAreaState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *ConfirmSelectAreaState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	if cb.Data == cbBack {
		r.Display(msgDutyManagement+msgSelectAction, taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}
	if strings.HasPrefix(cb.Data, "conf_area:") {
		return s.handleAreaSelection(ctx, cb, r)
	}
	return nil, nil
}

func (s *ConfirmSelectAreaState) handleAreaSelection(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	u, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	progressText := getMetProgress(ctx, s.cleaningUseCase, u.ID())
	areaID, _ := strconv.Atoi(strings.TrimPrefix(cb.Data, "conf_area:"))
	var filtered []ports.TaskViewModel
	for _, t := range s.tasks {
		if t.AreaID == areaID {
			filtered = append(filtered, t)
		}
	}
	r.Display(fmt.Sprintf("*%s*\n%s%s", filtered[0].AreaName, progressText, msgSelectConfirmTask), taskConfirmKeyboard(filtered))
	return NewConfirmTaskState(s.userUseCase, s.cleaningUseCase, filtered), nil
}

type ConfirmTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []ports.TaskViewModel
}

func NewConfirmTaskState(u ports.UserUseCase, c ports.CleaningUseCase, tasks []ports.TaskViewModel) *ConfirmTaskState {
	return &ConfirmTaskState{userUseCase: u, cleaningUseCase: c, tasks: tasks}
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
	progressText := getMetProgress(ctx, s.cleaningUseCase, user.ID())
	allAssigned, err := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
	if err != nil {
		return nil, err
	}
	r.Display(progressText+msgSelectArea, confirmAreaSelectKeyboard(allAssigned))
	return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, allAssigned), nil
}

func (s *ConfirmTaskState) handleCompletion(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder, user *user.User) (State, error) {
	taskID, _ := uuid.Parse(strings.TrimPrefix(cb.Data, "complete:"))
	if err := s.cleaningUseCase.CompleteTask(ctx, taskID, user.ID()); err != nil {
		r.Display(msgCompleteTaskErrPrefix+err.Error(), taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	areaID := s.tasks[0].AreaID
	areaName := s.tasks[0].AreaName
	allAssigned, _ := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
	var tasksInSameArea []ports.TaskViewModel
	for _, t := range allAssigned {
		if t.AreaID == areaID {
			tasksInSameArea = append(tasksInSameArea, t)
		}
	}

	progressText := getMetProgress(ctx, s.cleaningUseCase, user.ID())
	if len(tasksInSameArea) > 0 {
		msg := fmt.Sprintf("*%s*\n%s%s", areaName, progressText, msgTaskCompleted+msgTakeNextConfirmTask)
		r.Display(msg, taskConfirmKeyboard(tasksInSameArea))
		return s, nil
	}
	if len(allAssigned) > 0 {
		msg := fmt.Sprintf("%s%s", progressText, fmt.Sprintf(msgConfirmAreaFinished, areaName))
		r.Display(msg, confirmAreaSelectKeyboard(allAssigned))
		return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, allAssigned), nil
	}
	r.Display(msgConfirmAllFinished, taskMenuKeyboard())
	return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
}
