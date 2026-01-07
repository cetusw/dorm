package telegram

import (
	"context"
	"dorm/pkg/core/ports"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
		r.Display(msgRegisterErr, nil)
		return nil, err
	}
	r.SendMenu(msgRegisterSuccess, mainKeyboard())
	return NewMainMenuState(s.userUseCase, s.cleaningUseCase), nil
}
func (s *RegistrationState) HandleCallback(_ context.Context, _ *tgbotapi.CallbackQuery, _ *Responder) (State, error) {
	return nil, nil
}
