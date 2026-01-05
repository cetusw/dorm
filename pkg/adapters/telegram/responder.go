package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Responder struct {
	api           *tgbotapi.BotAPI
	chatID        int64
	lastBotMsgID  int
	callbackMsgID int
	onSent        func(int)
}

func (r *Responder) Display(text string, kb interface{}) {
	if r.callbackMsgID != 0 {
		edit := tgbotapi.NewEditMessageTextAndMarkup(r.chatID, r.callbackMsgID, text, tgbotapi.InlineKeyboardMarkup{})
		edit.ParseMode = "Markdown"
		if kb != nil {
			if inline, ok := kb.(tgbotapi.InlineKeyboardMarkup); ok {
				edit.ReplyMarkup = &inline
			}
		}
		sentMsg, err := r.api.Send(edit)
		if err == nil {
			if r.onSent != nil {
				r.onSent(sentMsg.MessageID)
			}
			return
		}
	}

	if r.lastBotMsgID != 0 {
		r.api.Send(tgbotapi.NewDeleteMessage(r.chatID, r.lastBotMsgID))
	}

	msg := tgbotapi.NewMessage(r.chatID, text)
	msg.ReplyMarkup = kb
	msg.ParseMode = "Markdown"
	sentMsg, err := r.api.Send(msg)

	if err == nil && r.onSent != nil {
		r.onSent(sentMsg.MessageID)
	}
}

func (r *Responder) SendMenu(text string, kb tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(r.chatID, text)
	msg.ReplyMarkup = kb
	m, err := r.api.Send(msg)
	r.track(m, err)
}

func (r *Responder) SendInline(text string, kb tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(r.chatID, text)
	msg.ReplyMarkup = kb
	m, err := r.api.Send(msg)
	r.track(m, err)
}

func (r *Responder) track(msg tgbotapi.Message, err error) {
	if err == nil && r.onSent != nil {
		r.onSent(msg.MessageID)
	}
}
