package telegram

import (
	"context"
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
	case "🧹 Задачи":
		r.SendInline("🛠 Управление дежурством\n\nВыберите действие ниже:", taskMenuKeyboard())
		return NewTaskMenuState(s.userUseCase, s.cleaningUseCase), nil
	case "📝 Мои задачи":
		user, _ := s.userUseCase.GetUserByTelegramID(ctx, msg.From.ID)
		tasks, _ := s.cleaningUseCase.GetAllAssignedTasks(ctx, user.ID())

		statsText := getFullProgress(ctx, s.cleaningUseCase, user)

		var b strings.Builder
		b.WriteString(statsText + "\n\n")
		b.WriteString("*Мои задачи:*\n\n")

		if len(tasks) == 0 {
			b.WriteString("У вас пока нет назначенных задач")
			r.Display(b.String(), mainKeyboard())
			return nil, nil
		}

		tasksByArea := make(map[string][]ports.TaskViewModel)
		for _, task := range tasks {
			tasksByArea[task.AreaName] = append(tasksByArea[task.AreaName], task)
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
				if task.IsDone {
					statusEmoji = "✅"
				}
				b.WriteString(fmt.Sprintf("%s %s\n", statusEmoji, task.Title))
			}
			b.WriteString("\n")
		}

		r.Display(b.String(), mainKeyboard())
		return nil, nil
	case "👤 Профиль":
		user, _ := s.userUseCase.GetUserByTelegramID(ctx, msg.From.ID)
		text := getFullProgress(ctx, s.cleaningUseCase, user)
		r.Display(text, mainKeyboard())
		return nil, nil
	}
	return nil, nil
}
func (s *MainMenuState) HandleCallback(_ context.Context, _ *tgbotapi.CallbackQuery, _ *Responder) (State, error) {
	return nil, nil
}
