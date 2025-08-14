package infrastructure

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram struct {
	Bot *tgbotapi.BotAPI
}

func NewTelegram(token string) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Telegram{Bot: bot}, nil
}

func (t *Telegram) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.Bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (t *Telegram) SendMessageWithReplyKeyboard(
	chatID int64,
	text string,
	buttons [][]string) {
	var rows [][]tgbotapi.KeyboardButton
	for _, buttonRow := range buttons {
		var row []tgbotapi.KeyboardButton
		for _, buttonText := range buttonRow {
			row = append(row, tgbotapi.NewKeyboardButton(buttonText))
		}
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	_, err := t.Bot.Send(msg)
	if err != nil {
	}
}

func (t *Telegram) GetUpdates(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return bot.GetUpdatesChan(u)
}
