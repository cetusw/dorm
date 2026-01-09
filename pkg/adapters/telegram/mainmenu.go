package telegram

import (
	"context"
	"dorm/pkg/core/ports/dto"
	"fmt"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dorm/pkg/core/ports"
)

type MainMenuState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
}

func NewMainMenuState(u ports.UserUseCase, c ports.CleaningUseCase) *MainMenuState {
	return &MainMenuState{userUseCase: u, cleaningUseCase: c}
}
func (s *MainMenuState) Name() string { return "MainMenu" }

func (s *MainMenuState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	switch msg.Text {
	case btnDuty:
		user, err := s.userUseCase.GetUserByTelegramID(ctx, msg.From.ID)
		if err != nil {
			return nil, err
		}
		if onDuty, err := s.cleaningUseCase.IsUserOnDuty(ctx, user.ID()); err != nil || !onDuty {
			r.Display(msgTeamNotOnDutyError, nil)
			return nil, err
		}

		r.SendInline(msgDutyManagement+msgDutyManagementDescription+msgSelectAction, taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	case btnMyTasks:
		return s.handleMyTasks(ctx, msg, r)
	case btnProfile:
		return s.handleProfile(ctx, msg, r)
	}
	return nil, nil
}

func (s *MainMenuState) handleMyTasks(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	user, _ := s.userUseCase.GetUserByTelegramID(ctx, msg.From.ID)
	tasks, _ := s.cleaningUseCase.GetAllAssignedTasks(ctx, user.ID())

	statsText := getFullProgress(ctx, s.cleaningUseCase, user)

	var b strings.Builder
	b.WriteString(statsText + msgMyTasksTitle)

	if len(tasks) == 0 {
		b.WriteString(msgNoAssignedTasks)
		r.Display(b.String(), mainKeyboard())
		return nil, nil
	}

	s.buildTasksList(&b, tasks)
	r.Display(b.String(), mainKeyboard())
	return nil, nil
}

func (s *MainMenuState) buildTasksList(b *strings.Builder, tasks []dto.TaskViewModel) {
	tasksByArea := make(map[string][]dto.TaskViewModel)
	for _, task := range tasks {
		key := fmt.Sprintf("%d этаж. %s", task.AreaFloor, task.AreaName)
		tasksByArea[key] = append(tasksByArea[key], task)
	}

	var sortedAreaNames []string
	for areaName := range tasksByArea {
		sortedAreaNames = append(sortedAreaNames, areaName)
	}
	sort.Strings(sortedAreaNames)

	for _, areaName := range sortedAreaNames {
		areaTasks := tasksByArea[areaName]
		SortTasks(areaTasks)
		b.WriteString(fmt.Sprintf("*%s*\n", areaName))
		for _, task := range areaTasks {
			statusEmoji := "📝"
			if task.IsCompleted {
				statusEmoji = "✅"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", statusEmoji, task.Title))
		}
		b.WriteString("\n")
	}
}

func (s *MainMenuState) handleProfile(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	user, _ := s.userUseCase.GetUserByTelegramID(ctx, msg.From.ID)
	profile, _ := s.userUseCase.GetUserProfile(ctx, user.ID())
	if profile == nil {
		r.Display(msgProfileLoadError, mainKeyboard())
		return nil, nil
	}

	onDuty, _ := s.cleaningUseCase.IsUserOnDuty(ctx, user.ID())

	dutyStatus := msgNotOnDutyTeam
	if onDuty {
		dutyStatus = msgOnDutyTeam
	}

	text := fmt.Sprintf(
		msgProfileFormat,
		profile.FirstName,
		profile.LastName,
		profile.RoomNumber,
		profile.DormitoryName,
		profile.GroupName,
		profile.TeamName,
		dutyStatus,
	)
	r.Display(text, mainKeyboard())
	return nil, nil
}

func (s *MainMenuState) HandleCallback(_ context.Context, _ *tgbotapi.CallbackQuery, _ *Responder) (State, error) {
	return nil, nil
}
