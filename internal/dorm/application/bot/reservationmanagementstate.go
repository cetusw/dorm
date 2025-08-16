package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ReservationManagementState struct{}

func (s *ReservationManagementState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	var text string
	var buttons [][]string
	var nextState State
	context.Telegram.ClearDialogue(chatID, update.Message.MessageID, context.LastMessageID)
	switch update.Message.Text {
	case message.Tasks:
		text = message.Tasks
		buttons = keyboard.TaskManagementMenu
		nextState = &TaskManagementState{}
	case message.Team:
		text = message.Team
		buttons = keyboard.TeamManagementMenu
		nextState = &TeamManagementState{}
		break
	case message.Payment:
		text = message.PaymentLink
		buttons = keyboard.BackMenu
		nextState = &PaymentManagementState{}
	case message.Profile:
		text = message.Profile // TODO: сделать профиль
		buttons = keyboard.BackMenu
		nextState = &ProfileManagementState{}
	default:
		//messageID, err := context.Telegram.SendMessageWithReplyKeyboard(
		//	chatID,
		//	message.Please,
		//	keyboard.MainMenu,
		//)
		//if err != nil || messageID == 0 {
		//	return
		//}
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

func (s *ReservationManagementState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons [][]string
	var nextState State
	switch update.CallbackQuery.Data {
	default:
		text = message.Back
		buttons = keyboard.MainMenu
		nextState = &MainState{}
	}
	context.Telegram.EditMessageTextAndKeyboard(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *ReservationManagementState) GetName() string {
	return "ReservationManagementState"
}
