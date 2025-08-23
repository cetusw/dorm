package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type baseState struct{}

func (s *baseState) AckCallbackAndChatID(context *Bot, update *tgbotapi.Update) int64 {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	return chatID
}

func (s *baseState) SendInlineAndGo(context *Bot, chatID int64, text string, buttons tgbotapi.InlineKeyboardMarkup, next State) {
	messageID, err := context.Telegram.SendMessageWithMarkup(chatID, text, buttons)
	if err != nil || messageID == 0 {
		return
	}
	context.LastMessageID = messageID
	context.SetState(next)
}

func (s *baseState) EditInlineAndGo(context *Bot, chatID int64, text string, buttons tgbotapi.InlineKeyboardMarkup, next State) {
	context.Telegram.EditMessageWithMarkup(chatID, context.LastMessageID, text, buttons)
	context.SetState(next)
}

func (s *baseState) SendReplyAndGo(context *Bot, chatID int64, text string, buttons tgbotapi.ReplyKeyboardMarkup, next State) {
	messageID, err := context.Telegram.SendMessageWithMarkup(chatID, text, buttons)
	if err != nil || messageID == 0 {
		return
	}
	context.LastMessageID = messageID
	context.SetState(next)
}

func (s *baseState) GetPointsSummary(
	context *Bot,
	teamID int,
	userID uuid.UUID,
	dutyID uuid.UUID,
) (userPoints int, pointsPerUser int, err error) {
	up, err := context.DutyTaskService.GetUserPointsByDutyID(userID, dutyID)
	if err != nil {
		return 0, 0, err
	}
	all, err := context.DutyTaskService.GetAllPointsByDuty(dutyID)
	if err != nil {
		return 0, 0, err
	}
	teamUsers, err := context.UserService.GetUsersByTeamID(teamID)
	if err != nil {
		return 0, 0, err
	}
	if len(teamUsers) == 0 {
		return up, 0, nil
	}
	return up, all / len(teamUsers), nil
}

func (s *baseState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	var text string
	context.Telegram.ClearDialogue(chatID, update.Message.MessageID, context.LastMessageID)

	switch update.Message.Text {
	case message.Tasks:
		text = message.TaskManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildTaskManagementKeyboard(), &TaskManagementState{})
		return
	case message.Team:
		text = message.TeamManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildTeamManagementKeyboard(), &TeamManagementState{})
		return
	case message.Payment:
		text = message.PaymentManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildBackKeyboard(), &PaymentManagementState{})
		return
	case message.Profile:
		text = message.Profile // TODO: сделать профиль
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildBackKeyboard(), &ProfileManagementState{})
		return
	default:
		messageID, err := context.Telegram.SendMessage(chatID, message.Please)
		if err != nil || messageID == 0 {
			return
		}
		context.LastMessageID = messageID
		return
	}
}

func (s *baseState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := s.AckCallbackAndChatID(context, update)
	s.SendReplyAndGo(context, chatID, message.MainState, keyboard.BuildMainStateKeyboard(), &MainState{})
}
