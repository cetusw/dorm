package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type baseState struct{}

func (s *baseState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	var text string
	var buttons [][]string
	var nextState State
	context.Telegram.ClearDialogue(chatID, update.Message.MessageID, context.LastMessageID)

	switch update.Message.Text {
	case message.Tasks:
		text = message.TaskManagementState
		buttons = keyboard.TaskManagementState
		nextState = &TaskManagementState{}
	case message.Team:
		text = message.TeamManagementState
		buttons = keyboard.TeamManagementState
		nextState = &TeamManagementState{}
	case message.Payment:
		text = message.PaymentManagementState
		buttons = keyboard.BackMenu // TODO: сделать меню или убрать его полностью
		nextState = &PaymentManagementState{}
	case message.Profile:
		text = message.Profile      // TODO: сделать профиль
		buttons = keyboard.BackMenu // TODO: сделать меню или убрать его полностью
		nextState = &ProfileManagementState{}
	default:
		messageID, err := context.Telegram.SendMessage(chatID, message.Please)
		if err != nil || messageID == 0 {
			return
		}
		context.LastMessageID = messageID
		return
	}

	messageId, err := context.Telegram.SendMessageWithInlineKeyboard(chatID, text, buttons)
	if err != nil || messageId == 0 {
		return
	}
	context.LastMessageID = messageId
	context.SetState(nextState)
}

func (s *baseState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")

	text := message.Back
	buttons := keyboard.MainState
	nextState := &MainState{}

	context.Telegram.EditMessageTextAndKeyboard(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}
