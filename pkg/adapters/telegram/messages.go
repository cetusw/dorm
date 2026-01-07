package telegram

const (
	btnProfile = "👤 Профиль"
	btnDuty    = "🛠 Дежурство"
	btnMyTasks = "📝 Мои задачи"

	btnAssign   = "➕ Взять задачу"
	btnUnassign = "➖ Отдать задачу"
	btnConfirm  = "✅ Подтвердить выполнение"
	btnBack     = "« Назад"
)

const (
	msgErrDefault                = "⚠️ Произошла ошибка. Пожалуйста, начните заново: /start"
	msgDutyManagement            = "🛠 Управление дежурством\n\n"
	msgDutyManagementDescription = "Здесь вы можете управлять своими задачами: брать, возвращать и подтверждать выполнение.\n"
	msgSelectAction              = "Выберите действие ниже:"
	msgOnDutyTeam                = "на дежурстве"
	msgNotOnDutyTeam             = "не на дежурстве"
	msgTeamNotOnDutyError        = "⛔️ Сейчас дежурит не ваша команда."

	msgProfileLoadError = "⚠️ Не удалось загрузить профиль. Возможно, вы еще не присоединены к команде."
	msgMyTasksTitle     = "*Мои задачи:*\n\n"
	msgNoAssignedTasks  = "У вас пока нет назначенных задач"
	msgProfileFormat    = "👤 %s %s\n\n🚪 Комната: %s\n🏢 Коливинг: %s\n👥 Группа: %s\n🛠 Команда: %s (%s)"

	msgNoFreeTasks         = "🎉 Свободных задач нет!"
	msgSelectArea          = "Выберите зону:"
	msgSelectTask          = "Выберите задачу:"
	msgTaskAssigned        = "✅ Задача взята!"
	msgTakeNextTask        = "Возьмите следующую:"
	msgAssignTaskErrPrefix = "❌ Ошибка: "

	msgNoActiveTasks         = "🤷‍♂️ У вас нет активных задач."
	msgSelectConfirmArea     = "Выберите зону, в которой вы выполнили задачу:"
	msgSelectConfirmTask     = "Выберите выполненную задачу:"
	msgTaskCompleted         = "✅ Задача выполнена!"
	msgTakeNextConfirmTask   = "Выберите следующую задачу:"
	msgNoTasksInArea         = "🎉 В зоне не осталось задач! Выберите следующую зону:"
	msgConfirmAllFinished    = "🎉 Поздравляю! Вы выполнили все свои задачи."
	msgCompleteTaskErrPrefix = "❌ Ошибка: "

	msgTaskUnassigned           = "✅ Задача отдана!"
	msgSelectNextTaskToUnassign = " Выберите следующую задачу:"
	msgAllTasksUnassigned       = "🎉 Вы отдали все свои задачи."
	msgUnassignTaskErrPrefix    = "❌ Ошибка: "

	msgWelcome         = "👋 Добро пожаловать!"
	msgReturnWelcome   = "👋 С возвращением!"
	msgAskForName      = "Введите Имя и Фамилию для регистрации:"
	msgRegisterSuccess = "✅ Регистрация успешна!"
	msgRegisterErr     = "⚠️ Ошибка регистрации. Попробуйте еще раз."
)
