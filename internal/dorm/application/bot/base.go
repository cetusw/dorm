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

func (s *baseState) GetCallbackID(callbackData string, prefix string) (int, error) {
	IDStr := strings.TrimPrefix(callbackData, prefix)
	ID, err := strconv.Atoi(IDStr)
	if err != nil {
		return 0, fmt.Errorf("ERROR: invalid ID in callback: %v", err)
	}

	return ID, nil
}

func (s *baseState) GetUnassignedTasks(context *Bot, areaID int) ([]model.DutyTaskView, error) {
	duty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return []model.DutyTaskView{}, err
	}
	tasks, err := context.DutyTaskService.GetUnassignedDutyTasksView(areaID, duty.DutyID)
	if err != nil {
		return []model.DutyTaskView{}, err
	}

	return tasks, nil
}

func (s *baseState) AssignTask(context *Bot, chatID int64, taskID uuid.UUID) error {
	duty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return err
	}
	user, err := context.UserService.GetUser(chatID)
	if err != nil {
		return err
	}
	return context.DutyTaskService.SetDutyTaskAssigneeID(&user.UserID, taskID, duty.DutyID)
}

func (s *baseState) UnassignTask(context *Bot, taskID uuid.UUID) error {
	duty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return err
	}
	return context.DutyTaskService.SetDutyTaskAssigneeID(nil, taskID, duty.DutyID)
}

func (s *baseState) GetUncompletedDutyTasksView(context *Bot, userID int64) ([]model.DutyTaskView, error) {
	user, err := context.UserService.GetUser(userID)
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

func (s *baseState) GetUnassignedAreas(context *Bot) ([]model.Area, error) {
	duty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return []model.Area{}, err
	}
	areas, err := context.AreaService.GetUnassignedAreasByDutyID(duty.DutyID)
	if err != nil {
		return []model.Area{}, err
	}
	return areas, nil
}

func (s *baseState) ConfirmTask(context *Bot, taskID uuid.UUID) error {
	duty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return err
	}
	err = context.DutyTaskService.CompleteDutyTaskByDutyIDAndTaskID(duty.DutyID, taskID)
	if err != nil {
		return err
	}

	return nil
}

func (s *baseState) ComputeUserProgress(context *Bot, chatID int64) (*model.UserProgress, error) {
	duty, err := context.DutyService.GetCurrentDuty()
	if err != nil {
		return &model.UserProgress{}, err
	}
	user, err := context.UserService.GetUser(chatID)
	if err != nil {
		return &model.UserProgress{}, err
	}
	userPoints, err := context.DutyTaskService.GetUserPoints(user.UserID, duty.DutyID)
	if err != nil {
		return &model.UserProgress{}, err
	}
	userConfirmedPoints, err := context.DutyTaskService.GetUserConfirmedPoints(user.UserID, duty.DutyID)
	if err != nil {
		return &model.UserProgress{}, err
	}
	dutyPoints, err := context.DutyTaskService.GetDutyPoints(duty.DutyID)
	if err != nil {
		return &model.UserProgress{}, err
	}
	userRequiredPoints, err := context.UserService.GetRequiredUserPoints(user.UserID, dutyPoints)
	if err != nil {
		return &model.UserProgress{}, err
	}
	return &model.UserProgress{
		UserPoints:          userPoints,
		UserConfirmedPoints: userConfirmedPoints,
		UserRequiredPoints:  userRequiredPoints,
	}, nil
}

func (s *baseState) Handle(context *Bot, update *tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	var text string
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)

	switch update.Message.Text {
	case message.StartCommand:
		s.SendReplyAndGo(context, chatID, message.MainState, keyboard.BuildMainStateKeyboard(), &MainState{})
	case message.Tasks:
		text = message.TaskManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildTaskManagementKeyboard(), &TaskManagementState{})
	case message.Payment:
		text = message.PaymentManagementState
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildBackKeyboard(), &PaymentManagementState{})
	case message.Profile:
		user, err := context.UserService.GetUser(chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		duty, err := context.DutyService.GetCurrentDuty()
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		teamMembers, err := context.UserService.GetUsersByTeamID(*user.TeamID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		progress, err := s.ComputeUserProgress(context, chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		text, err := message.BuildProfileText(*user, *duty, teamMembers, *progress)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		s.SendInlineAndGo(context, chatID, text, keyboard.BuildBackKeyboard(), &ProfileInfoState{})
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
