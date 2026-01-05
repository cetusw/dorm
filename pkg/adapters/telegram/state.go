package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type State interface {
	HandleMessage(ctx context.Context, msg *tgbotapi.Message, r *Responder) (State, error)
	HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, r *Responder) (State, error)
	Name() string
}
