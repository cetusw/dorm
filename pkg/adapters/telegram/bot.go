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

// TODO: refactor
func (b *BotAdapter) processUpdate(update tgbotapi.Update) {
	var chatID, userID int64
	var callbackMsgID int

	if update.Message != nil {
		chatID = update.Message.Chat.ID
		userID = update.Message.From.ID
		b.api.Send(tgbotapi.NewDeleteMessage(chatID, update.Message.MessageID))
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		userID = update.CallbackQuery.From.ID
		callbackMsgID = update.CallbackQuery.Message.MessageID
		b.api.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
	} else {
		return
	}

	state := b.getState(userID)
	if state == nil {
		state = NewAuthState(b.userUseCase, b.cleaningUseCase)
	}

	b.muMsg.Lock()
	lastID := b.lastBotMsg[userID]
	b.muMsg.Unlock()

	r := &Responder{
		api:           b.api,
		chatID:        chatID,
		lastBotMsgID:  lastID,
		callbackMsgID: callbackMsgID,
		onSent: func(id int) {
			b.muMsg.Lock()
			b.lastBotMsg[userID] = id
			b.muMsg.Unlock()
		},
	}
	ctx := context.Background()
	var nextState State
	var err error

	if update.Message != nil {
		nextState, err = state.HandleMessage(ctx, update.Message, r)
	} else if update.CallbackQuery != nil {
		b.api.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		nextState, err = state.HandleCallback(ctx, update.CallbackQuery, r)
	}

	if err != nil {
		log.Printf("Bot Error: %v", err)
		r.Display("⚠️ Произошла ошибка. Пожалуйста, начните заново: /start", nil)
		b.setState(userID, nil)
		return
	}

	if nextState != nil {
		b.setState(userID, nextState)
	}
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
