package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"dorm/internal/dorm/application/model"
	"fmt"
	"strconv"
	"strings"

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

func (s *AreaSelectionState) GetAreaID(callbackData string) (int, error) {
	areaIDStr := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixArea)
	areaID, err := strconv.Atoi(areaIDStr)
	if err != nil {
		return 0, fmt.Errorf("ERROR: invalid area ID in callback: %v", err)
	}

	return areaID, nil
}

func (s *AreaSelectionState) GetCurrentDuty(context *Bot) (*model.Duty, error) {
	duty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get last duty: %v", err)
	}

	return duty, nil
}

func (s *AreaSelectionState) GetUnassignedTasks(
	context *Bot,
	areaID int,
	dutyID uuid.UUID,
) ([]model.DutyTaskView, error) {
	tasks, err := context.DutyTaskService.GetUnassignedDutyTasksView(areaID, dutyID)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get unassigned tasks for area %d: %v", areaID, err)
	}

	return tasks, nil
}

func (s *AreaSelectionState) GetUser(context *Bot, chatID int64) (*model.User, error) {
	user, err := context.UserService.GetUser(chatID)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get user while getting duty tasks: %v", err)
	}

	return user, nil
}

func (s *baseState) GetUncompletedDutyTasksView(context *Bot, update *tgbotapi.Update) ([]model.DutyTaskView, error) {
	user, err := context.UserService.GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get user while getting duty tasks: %v", err)
	}
	lastDuty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get last duty tasks: %v", err)
	}
	dutyTasksReadable, err := context.DutyTaskService.GetUncompletedDutyTasksView(user.UserID, lastDuty.DutyID)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get areas for keyboard: %v", err)
	}

	return dutyTasksReadable, nil
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
	all, err := context.DutyTaskService.GetAllPointsByDutyID(dutyID)
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

func (s *baseState) Handle(context *Bot, update *tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	var text string
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)

	switch update.Message.Text {
	case message.Tasks:
		text = message.TaskManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildTaskManagementKeyboard(), &TaskManagementState{})
	case message.Team:
		text = message.TeamManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildTeamManagementKeyboard(), &TeamManagementState{})
	case message.Payment:
		text = message.PaymentManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildBackKeyboard(), &PaymentManagementState{})
	case message.Profile:
		text = message.Profile // TODO: сделать профиль
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildBackKeyboard(), &ProfileManagementState{})
	default:
		messageID, err := context.Telegram.SendMessage(chatID, message.Please)
		if err != nil || messageID == 0 {
			return err
		}
		context.LastMessageID = messageID
	}

	return nil
}

func (s *baseState) HandleCallback(context *Bot, update *tgbotapi.Update) error {
	chatID := s.AckCallbackAndChatID(context, update)
	s.SendReplyAndGo(context, chatID, message.MainState, keyboard.BuildMainStateKeyboard(), &MainState{})
	return nil
}
