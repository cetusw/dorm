package message

import (
	"dorm/internal/dorm/application/model"
	"fmt"
	"strings"
)

const (
	MainState = "👋 Добро пожаловать в бот дежурств!\n\nЧто можно делать:\n• ➕ Взять задачу — выберите зону и задание\n• ✅ Подтвердить выполнение — отметьте завершённые задачи\n• ↩️ Снять назначение — если передумали или не успеваете\n\nБаллы начисляются за подтверждённые задачи. Наберите достаточно подтверждённых баллов за дежурство, чтобы выполнить норму.\n\nИспользуйте меню для навигации."
	Please    = "ℹ️ Пожалуйста, пользуйтесь меню ниже.\nЕсли его не видно — нажмите «Меню» в поле ввода."

	Registration           = "📝 Давайте познакомимся!\nОтправьте ваши ФИО одним сообщением. Примеры:\n• Иван Иванов\n• Иван Иванов Иванович"
	RegistrationSuccess    = "✅ Готово! Вы успешно зарегистрированы.\nОткрылось главное меню — можете выбрать раздел."
	RegistrationFail       = "⚠️ Не удалось сохранить данные. Попробуйте ещё раз чуть позже."
	PaymentManagementState = "💳 Оплата проживания\nПерейдите по ссылке для оплаты:\n\nhttps://clck.ru/3Nq23U\n\nПосле оплаты вернитесь в бот."
	ProfileHeader          = "👤 Профиль\n"

	TaskManagementState   = "🧹 Управление задачами."
	NotOnDutyTeam         = "✅ Твоя команда не дежурит на этой неделе, брать задачи ни к чему."
	AreaSelectionState    = "📍 Выберите зону для уборки.\n\n" + PointsSummary
	ConfirmExecutionState = "✅ Что подтвердим?\nВыберите задачу, которую уже выполнили.\n\n" + PointsSummary
	TaskAssignmentState   = "🧩 Выберите задачу, которую готовы выполнить.\n\n" + PointsSummary
	TaskUnassignmentState = "↩️ Выберите задачу, которую вы не будете выполнять.\n\n" + PointsSummary

	NoTasksToConfirm  = "ℹ️ Пока нет задач для подтверждения. Сначала возьмите задачу в работу."
	AllTasksConfirmed = "🎉 Все задачи подтверждены — отлично сработано!"

	NoCurrentDuty = "⏳ Сейчас нет активного дежурства. Загляните позже."
	NoDutyTasks   = "🗒️ Пока нет задач для текущего дежурства. Проверьте позже."

	PointsSummary = "Ваши текущие баллы:\n• Взято: %d\n• Подтверждено: %d\n• Нужно подтвердить: %.1f"
)

const (
	Error = "Произошла ошибка при выполнении действия. Пожалуйста, повторите позже"
)

func BuildFullName(user model.User) string {
	var nameParts []string
	if user.LastName != "" {
		nameParts = append(nameParts, user.LastName)
	}
	if user.FirstName != "" {
		nameParts = append(nameParts, user.FirstName)
	}
	if user.MiddleName != nil && *user.MiddleName != "" {
		nameParts = append(nameParts, *user.MiddleName)
	}

	return strings.TrimSpace(strings.Join(nameParts, " "))
}

func BuildTeamInfo(teamMembers []model.User) string {
	var membersBuilder strings.Builder
	for i, m := range teamMembers {
		var parts []string
		if m.LastName != "" {
			parts = append(parts, m.LastName)
		}
		if m.FirstName != "" {
			parts = append(parts, m.FirstName)
		}
		if m.MiddleName != nil && *m.MiddleName != "" {
			parts = append(parts, *m.MiddleName)
		}
		name := strings.TrimSpace(strings.Join(parts, " "))
		if name == "" {
			name = fmt.Sprintf("Участник #%d", i+1)
		}
		membersBuilder.WriteString("• ")
		membersBuilder.WriteString(name)
		membersBuilder.WriteString("\n")
	}

	teamBlock := "Участники команды:\n"
	if membersBuilder.Len() == 0 {
		teamBlock += "• нет данных\n"
	} else {
		teamBlock += membersBuilder.String()
	}

	return teamBlock
}

func BuildTeamStatus(user model.User, duty *model.Duty) string {
	teamStatus := "Команда: не определена"
	if user.TeamID != nil {
		if duty != nil && duty.TeamID == *user.TeamID {
			teamStatus = "Команда: дежурная на этой неделе ❗"
		} else {
			teamStatus = "Команда: не дежурит на этой неделе ✅"
		}
	}

	return teamStatus
}

func BuildPointsInfo(user model.User, duty *model.Duty, progress model.UserProgress) string {
	pointsBlock := ""
	if duty != nil && duty.TeamID == *user.TeamID {
		pointsBlock = "\n" + fmt.Sprintf(PointsSummary, progress.UserPoints, progress.UserConfirmedPoints, progress.UserRequiredPoints)
	}

	return pointsBlock
}

func BuildProfileText(
	user model.User,
	duty *model.Duty,
	teamMembers []model.User,
	progress model.UserProgress,
) (string, error) {
	fullName := BuildFullName(user)
	teamBlock := BuildTeamInfo(teamMembers)
	teamStatus := BuildTeamStatus(user, duty)
	pointsBlock := BuildPointsInfo(user, duty, progress)

	return fmt.Sprintf("%s\n%s\n%s\n%s\n\n%s", ProfileHeader, fullName, teamStatus, pointsBlock, teamBlock), nil
}

// TODO: разобраться с переносами строк при отсутствии блока баллов
