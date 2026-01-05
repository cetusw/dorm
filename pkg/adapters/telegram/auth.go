package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dorm/pkg/core/ports"
)

type AuthState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
}

func NewAuthState(u ports.UserUseCase, c ports.CleaningUseCase) *AuthState {
	return &AuthState{userUseCase: u, cleaningUseCase: c}
}
func (s *AuthState) Name() string { return "Auth" }

func (s *AuthState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	user, err := s.userUseCase.GetUserByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return nil, err
	}

	if user != nil {
		r.SendMenu("👋 С возвращением!", mainKeyboard())
		return NewMainMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	r.Display("👋 Добро пожаловать! Введите Имя и Фамилию для регистрации:", nil)
	return NewRegistrationState(s.userUseCase, s.cleaningUseCase), nil
}
func (s *AuthState) HandleCallback(_ context.Context, _ *tgbotapi.CallbackQuery, _ *Responder) (State, error) {
	return nil, nil
}

type RegistrationState struct {
	userUseCase     ports.UserUseCase
	cleaningUseCase ports.CleaningUseCase
}

func NewRegistrationState(u ports.UserUseCase, c ports.CleaningUseCase) *RegistrationState {
	return &RegistrationState{userUseCase: u, cleaningUseCase: c}
}
func (s *RegistrationState) Name() string { return "Registration" }

func (s *RegistrationState) HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error) {
	if err := s.userUseCase.RegisterUser(ctx, msg.From.ID, msg.Text); err != nil {
		r.Display("⚠️ Ошибка регистрации. Попробуйте еще раз.", nil)
		return nil, err
	}
	r.SendMenu("✅ Регистрация успешна!", mainKeyboard())
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase), nil
}
func (s *RegistrationState) HandleCallback(_ context.Context, _ *tgbotapi.CallbackQuery, _ *Responder) (State, error) {
	return nil, nil
}
