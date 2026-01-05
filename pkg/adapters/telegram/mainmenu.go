package telegram

import (
	"context"
	"fmt"

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

	case "👤 Профиль":
		user, _ := s.userUseCase.GetUserByTelegramID(ctx, msg.From.ID)
		stats, _ := s.cleaningUseCase.GetUserStats(ctx, user.ID())

		status := "🟢 Норма выполнена"
		if !stats.IsQuotaMet() {
			status = "🔴 Норма не выполнена"
		}

		text := fmt.Sprintf(
			"👤 %s %s\n\n📊 Баллы: %d\n🎯 Цель: %.1f\n%s",
			user.FirstName(), user.LastName(),
			stats.ConfirmedPoints, stats.RequiredPoints,
			status,
		)
		r.Display(text, mainKeyboard())
		return nil, nil
	}
	return nil, nil
}
func (s *MainMenuState) HandleCallback(_ context.Context, _ *tgbotapi.CallbackQuery, _ *Responder) (State, error) {
	return nil, nil
}
