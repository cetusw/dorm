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
		return 0, err
	}

	return sentMsg.MessageID, nil
}

func (t *Telegram) SendMessageWithMarkup(
	chatID int64,
	text string,
	keyboard interface{},
) (int, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	sentMsg, err := t.Bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message with markup: %v", err)
		return 0, err
	}
	return sentMsg.MessageID, nil
}

func (t *Telegram) EditMessageWithMarkup(
	chatID int64,
	messageID int,
	text string,
	keyboard tgbotapi.InlineKeyboardMarkup,
) {
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, keyboard)

	_, err := t.Bot.Send(editMsg)
	if err != nil {
		log.Printf("Failed to edit message: %v", err)
	}
}

func (t *Telegram) AnswerCallbackQuery(callbackQueryID string, text string) {
	config := tgbotapi.NewCallback(callbackQueryID, text)
	_, err := t.Bot.Request(config)
	if err != nil {
		log.Printf("Failed to answer callback query: %v", err)
	}
}

func (t *Telegram) DeleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := t.Bot.Request(deleteMsg)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}
}

func (t *Telegram) ClearDialogue(chatID int64, firstMessageID int, secondMessageID int) {
	t.DeleteMessage(chatID, firstMessageID)
	t.DeleteMessage(chatID, secondMessageID)
}

func (t *Telegram) GetUpdates(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return bot.GetUpdatesChan(u)
}
