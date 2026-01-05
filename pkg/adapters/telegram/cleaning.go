package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"

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
	user, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)

	onDuty, err := s.cleaningUseCase.IsUserOnDuty(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	if !onDuty {
		r.Display("⛔️ Сейчас дежурит не ваша команда. Вы не можете брать или выполнять задачи.", nil)
		return NewMainMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	switch cb.Data {
	case cbAssign:
		progressText := getSimplifiedProgress(ctx, s.cleaningUseCase, user.ID())
		tasks, _ := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())
		if len(tasks) == 0 {
			r.Display("🎉 Свободных задач нет!", taskMenuKeyboard())
			return s, nil
		}
		r.Display(progressText+"📍 Выберите зону:", areaSelectKeyboard(tasks))
		return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil

	case cbConfirm:
		tasks, _ := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
		if len(tasks) == 0 {
			r.Display("🤷‍♂️ У вас нет активных задач.", taskMenuKeyboard())
			return s, nil
		}
		kb := confirmAreaSelectKeyboard(tasks)
		r.Display("✅ В какой зоне вы завершили уборку?", kb)
		return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
	}
	return nil, nil
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
		r.Display("🛠 Управление дежурством\n\nВыберите действие ниже:", taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	if strings.HasPrefix(cb.Data, "area:") {
		user, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
		progressText := getSimplifiedProgress(ctx, s.cleaningUseCase, user.ID())
		areaID, _ := strconv.Atoi(strings.TrimPrefix(cb.Data, "area:"))

		var filtered []ports.TaskViewModel
		for _, t := range s.tasks {
			if t.AreaID == areaID {
				filtered = append(filtered, t)
			}
		}

		if len(filtered) == 0 {
			r.Display(progressText+"⚠️ В этой зоне задач не осталось. Выберите другую:", areaSelectKeyboard(s.tasks))
			return s, nil
		}

		r.Display(progressText+"👇 Выберите задачу:", taskSelectKeyboard(filtered))
		return NewSelectTaskState(s.userUseCase, s.cleaningUseCase, areaID), nil
	}

	return nil, nil
}

type SelectTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	areaID          int
}

func NewSelectTaskState(
	u ports.UserUseCase,
	c ports.CleaningUseCase,
	areaID int,
) *SelectTaskState {
	return &SelectTaskState{
		userUseCase:     u,
		cleaningUseCase: c,
		areaID:          areaID,
	}
}

func (s *SelectTaskState) Name() string { return "SelectTask" }

func (s *SelectTaskState) HandleMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
	r *Responder,
) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *SelectTaskState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	user, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)
	if cb.Data == cbBack {
		tasks, _ := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())
		r.Display("📍 Выберите зону:", areaSelectKeyboard(tasks))
		return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, tasks), nil
	}

	if strings.HasPrefix(cb.Data, "assign:") {
		taskID, _ := uuid.Parse(strings.TrimPrefix(cb.Data, "assign:"))

		if err := s.cleaningUseCase.AssignTask(ctx, taskID, user.ID()); err != nil {
			r.Display("❌ Ошибка: "+err.Error(), taskMenuKeyboard())
			return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
		}

		progressText := getSimplifiedProgress(ctx, s.cleaningUseCase, user.ID())

		allTasks, _ := s.cleaningUseCase.GetTaskCandidates(ctx, user.ID())

		var tasksInSameArea []ports.TaskViewModel
		for _, t := range allTasks {
			if t.AreaID == s.areaID {
				tasksInSameArea = append(tasksInSameArea, t)
			}
		}

		if len(tasksInSameArea) > 0 {
			r.Display(progressText+"✅ Задача взята! Возьмите следующую:", taskSelectKeyboard(tasksInSameArea))
			return s, nil
		}

		if len(allTasks) > 0 {
			r.Display(progressText+"✅ В этой зоне задач не осталось. Выберите другую:", areaSelectKeyboard(allTasks))
			return NewSelectAreaState(s.userUseCase, s.cleaningUseCase, allTasks), nil
		}

		r.Display(progressText+"✅ Задача взята! Больше свободных задач нет.", taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	return nil, nil
}

type ConfirmSelectAreaState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	tasks           []ports.TaskViewModel
}

func NewConfirmSelectAreaState(
	u ports.UserUseCase,
	c ports.CleaningUseCase,
	t []ports.TaskViewModel,
) *ConfirmSelectAreaState {
	return &ConfirmSelectAreaState{userUseCase: u, cleaningUseCase: c, tasks: t}
}

func (s *ConfirmSelectAreaState) Name() string { return "ConfirmSelectArea" }

func (s *ConfirmSelectAreaState) HandleMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
	r *Responder,
) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *ConfirmSelectAreaState) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error) {
	if cb.Data == cbBack {
		r.Display("Управление дежурством:", taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	if strings.HasPrefix(cb.Data, "conf_area:") {
		areaID, _ := strconv.Atoi(strings.TrimPrefix(cb.Data, "conf_area:"))

		var filtered []ports.TaskViewModel
		for _, t := range s.tasks {
			if t.AreaID == areaID {
				filtered = append(filtered, t)
			}
		}

		r.Display("✅ Выберите выполненную задачу:", taskConfirmKeyboard(filtered))
		return NewConfirmTaskState(s.userUseCase, s.cleaningUseCase, areaID), nil
	}
	return nil, nil
}

type ConfirmTaskState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
	areaID          int
}

func NewConfirmTaskState(u ports.UserUseCase, c ports.CleaningUseCase, areaID int) *ConfirmTaskState {
	return &ConfirmTaskState{userUseCase: u, cleaningUseCase: c, areaID: areaID}
}

func (s *ConfirmTaskState) Name() string { return "ConfirmTask" }

func (s *ConfirmTaskState) HandleMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
	r *Responder,
) (State, error) {
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase).HandleMessage(ctx, msg, r)
}

func (s *ConfirmTaskState) HandleCallback(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	r *Responder,
) (State, error) {
	user, _ := s.userUseCase.GetUserByTelegramID(ctx, cb.From.ID)

	if cb.Data == cbBack {
		allAssigned, _ := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())
		r.Display("✅ Выберите зону:", confirmAreaSelectKeyboard(allAssigned))
		return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, allAssigned), nil
	}

	if strings.HasPrefix(cb.Data, "complete:") {
		taskID, _ := uuid.Parse(strings.TrimPrefix(cb.Data, "complete:"))

		if err := s.cleaningUseCase.CompleteTask(ctx, taskID, user.ID()); err != nil {
			r.Display("❌ Ошибка: "+err.Error(), taskMenuKeyboard())
			return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
		}

		allAssigned, _ := s.cleaningUseCase.GetUncompletedAssignedTasks(ctx, user.ID())

		var tasksInSameArea []ports.TaskViewModel
		for _, t := range allAssigned {
			if t.AreaID == s.areaID {
				tasksInSameArea = append(tasksInSameArea, t)
			}
		}

		if len(tasksInSameArea) > 0 {
			r.Display("🎉 Отлично! Еще что-то в этой зоне?", taskConfirmKeyboard(tasksInSameArea))
			return s, nil
		}

		if len(allAssigned) > 0 {
			r.Display("👏 В этой зоне всё готово! Выберите следующую:", confirmAreaSelectKeyboard(allAssigned))
			return NewConfirmSelectAreaState(s.userUseCase, s.cleaningUseCase, allAssigned), nil
		}

		r.Display("🥳 Поздравляю! Вы выполнили все свои задачи.", taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	}
	return nil, nil
}
