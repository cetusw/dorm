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

func (t *Telegram) SendMessage(chatID int64, text string) (int, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	sentMsg, err := t.Bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return 0, err
	}

	return sentMsg.MessageID, err
}

func (t *Telegram) SendMessageWithReplyKeyboard(
	chatID int64,
	text string,
	buttons [][]string) (int, error) {
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

	sentMsg, err := t.Bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message with reply keyboard: %v", err)
		return 0, err
	}

	return sentMsg.MessageID, nil
}

func (t *Telegram) DeleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := t.Bot.Request(deleteMsg)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}
}

func (t *Telegram) GetUpdates(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return bot.GetUpdatesChan(u)
}
