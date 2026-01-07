package telegram

const (
	btnProfile = "👤 Профиль"
	btnDuty    = "🛠 Дежурство"
	btnMyTasks = "📝 Мои задачи"
	btnBack    = "« Назад"
)

const (
	msgErrDefault                = "⚠️ Произошла ошибка. Пожалуйста, начните заново: /start"
	msgDutyManagement            = "🛠 Управление дежурством\n\n"
	msgDutyManagementDescription = "Здесь вы можете взять задачу и подтвердить её выполнение.\n"
	msgSelectAction              = "Выберите действие ниже:"
	msgOnDutyTeam                = "на дежурстве"
	msgNotOnDutyTeam             = "не на дежурстве"
	msgTeamNotOnDutyError        = "⛔️ Сейчас дежурит не ваша команда."

	msgProfileLoadError = "⚠️ Не удалось загрузить профиль. Возможно, вы еще не присоединены к команде."
	msgMyTasksTitle     = "*Мои задачи:*\n\n"
	msgNoAssignedTasks  = "У вас пока нет назначенных задач"
	msgProfileFormat    = "👤 %s %s\n\n🚪 Комната: %s\n🏢 Коливинг: %s\n👥 Группа: %s\n🛠 Команда: %s (%s)"

	msgNoFreeTasks         = "🎉 Свободных задач нет!"
	msgSelectArea          = "📍 Выберите зону:"
	msgNoTasksInArea       = "🎉 В этой зоне задач не осталось. Выберите другую зону:"
	msgSelectTask          = "Выберите задачу:"
	msgTaskAssigned        = "✅ Задача взята!"
	msgTakeNextTask        = " Возьмите следующую:"
	msgNoMoreTasksInArea   = "\n🎉 В этой зоне задач не осталось.\nВыберите другую зону:"
	msgNoMoreFreeTasks     = "🎉 Больше свободных задач нет."
	msgAssignTaskErrPrefix = "❌ Ошибка: "

	msgNoActiveTasks         = "🤷‍♂️ У вас нет активных задач."
	msgSelectConfirmArea     = "Выберите зону, в которой вы выполнили задачу:"
	msgSelectConfirmTask     = "Выберите выполненную задачу:"
	msgTaskCompleted         = "✅ Задача выполнена!"
	msgTakeNextConfirmTask   = " Выберити следующую задачу:"
	msgConfirmAreaFinished   = "🎉 В зоне \"%s\" всё готово! Выберити следующую зону:"
	msgConfirmAllFinished    = "🎉 Поздравляю! Вы выполнили все свои задачи."
	msgCompleteTaskErrPrefix = "❌ Ошибка: "

	msgWelcome         = "👋 Добро пожаловать!"
	msgReturnWelcome   = "👋 С возвращением!"
	msgAskForName      = "Введите Имя и Фамилию для регистрации:"
	msgRegisterSuccess = "✅ Регистрация успешна!"
	msgRegisterErr     = "⚠️ Ошибка регистрации. Попробуйте еще раз."
)
