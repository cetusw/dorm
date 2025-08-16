package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ConfirmationManagementState struct{}

func (s *ConfirmationManagementState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := ""
	buttons := keyboard.MainMenu
	context.Telegram.DeleteMessage(chatID, context.LastMessageID)
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
	switch update.Message.Text {
	default:
		text = message.Back
		buttons = keyboard.MainMenu
		context.SetState(&MainState{})
	}
	messageId, err := context.Telegram.SendMessageWithInlineKeyboard(chatID, text, buttons)
	if err != nil || messageId == 0 {
		return
	}
	context.LastMessageID = messageId
}

func (s *ConfirmationManagementState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons [][]string
	var nextState State
	switch update.Message.Text {
	default:
		text = message.Back
		buttons = keyboard.MainMenu
		nextState = &MainState{}
	}
	context.Telegram.EditMessageTextAndKeyboard(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *ConfirmationManagementState) GetName() string {
	return "ConfirmationManagementState"
}
