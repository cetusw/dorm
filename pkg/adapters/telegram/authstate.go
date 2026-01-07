package telegram

import (
	"context"
	"dorm/pkg/core/ports"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
		r.SendMenu(msgReturnWelcome, mainKeyboard())
		return NewMainMenuState(s.userUseCase, s.cleaningUseCase), nil
	}

	r.Display(msgWelcome+"\n"+msgAskForName, nil)
	return NewRegistrationState(s.userUseCase, s.cleaningUseCase), nil
}
func (s *AuthState) HandleCallback(_ context.Context, _ *tgbotapi.CallbackQuery, _ *Responder) (State, error) {
	return nil, nil
}
