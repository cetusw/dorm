package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MainState struct{}

func (s *MainState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	var text string
	var buttons [][]string
	var nextState State
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
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

func (s *MainState) HandleCallback(context *Bot, update *tgbotapi.Update) {}

func (s *MainState) GetName() string {
	return "MainState"
}
