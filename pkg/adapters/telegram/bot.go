package telegram

import (
	"context"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dorm/pkg/core/ports"
)

type BotAdapter struct {
	api             *tgbotapi.BotAPI
	cleaningUseCase ports.CleaningUseCase
	userUseCase     ports.UserUseCase
	states          map[int64]State
	statesMu        sync.RWMutex
	lastBotMsg      map[int64]int
	muMsg           sync.Mutex
}

func NewBotAdapter(
	token string,
	cleaningUC ports.CleaningUseCase,
	userUC ports.UserUseCase,
) (*BotAdapter, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &BotAdapter{
		api:             api,
		cleaningUseCase: cleaningUC,
		userUseCase:     userUC,
		states:          make(map[int64]State),
		lastBotMsg:      make(map[int64]int),
	}, nil
}

func (b *BotAdapter) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	log.Println("Telegram Bot started")

	for update := range updates {
		go b.processUpdate(update)
	}
}

type updateInfo struct {
	ChatID        int64
	UserID        int64
	CallbackMsgID int
}

func (b *BotAdapter) processUpdate(update tgbotapi.Update) {
	info, ok := b.extractUpdateInfo(update)
	if !ok {
		return
	}
	b.handleRequest(context.Background(), update, info)
}

func (b *BotAdapter) handleRequest(ctx context.Context, update tgbotapi.Update, info *updateInfo) {
	state := b.getOrCreateState(info.UserID)
	responder := b.createResponder(info)

	nextState, err := b.executeStateHandler(ctx, state, update, responder)
	if err != nil {
		log.Printf("Bot Error: %v", err)
		responder.Display(msgErrDefault, nil)
		b.setState(info.UserID, nil)
		return
	}

	if nextState != nil {
		b.setState(info.UserID, nextState)
	}
}

func (b *BotAdapter) extractUpdateInfo(update tgbotapi.Update) (*updateInfo, bool) {
	if update.Message != nil {
		b.api.Send(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
		return &updateInfo{
			ChatID: update.Message.Chat.ID,
			UserID: update.Message.From.ID,
		}, true
	}

	if update.CallbackQuery != nil {
		b.api.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		return &updateInfo{
			ChatID:        update.CallbackQuery.Message.Chat.ID,
			UserID:        update.CallbackQuery.From.ID,
			CallbackMsgID: update.CallbackQuery.Message.MessageID,
		}, true
	}

	return nil, false
}

func (b *BotAdapter) getOrCreateState(userID int64) State {
	if state := b.getState(userID); state != nil {
		return state
	}
	return NewAuthState(b.userUseCase, b.cleaningUseCase)
}

func (b *BotAdapter) createResponder(info *updateInfo) *Responder {
	b.muMsg.Lock()
	lastID := b.lastBotMsg[info.UserID]
	b.muMsg.Unlock()

	return &Responder{
		api:           b.api,
		chatID:        info.ChatID,
		lastBotMsgID:  lastID,
		callbackMsgID: info.CallbackMsgID,
		onSent: func(id int) {
			b.muMsg.Lock()
			b.lastBotMsg[info.UserID] = id
			b.muMsg.Unlock()
		},
	}
}

func (b *BotAdapter) executeStateHandler(ctx context.Context, state State, update tgbotapi.Update, r *Responder) (State, error) {
	if update.Message != nil {
		return state.HandleMessage(ctx, update.Message, r)
	}
	return state.HandleCallback(ctx, update.CallbackQuery, r)
}

func (b *BotAdapter) getState(userID int64) State {
	b.statesMu.RLock()
	defer b.statesMu.RUnlock()
	return b.states[userID]
}

func (b *BotAdapter) setState(userID int64, state State) {
	b.statesMu.Lock()
	defer b.statesMu.Unlock()
	if state == nil {
		delete(b.states, userID)
	} else {
		b.states[userID] = state
	}
}
