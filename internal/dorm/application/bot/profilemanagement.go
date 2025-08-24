package bot

type ProfileManagementState struct {
	baseState
}

// TODO: дописать логику профиля. Нужно отображать информацию о пользователе
// - ФИО
// - Команда
// - Дата следующего дежурства
// - Количество набранных баллов за все уборки (по приколу)
// - Оплачен ли коливинг за текущий месяц (опционально, но если получиться придумать как это сделать, то было бы круто)

func (s *ProfileManagementState) GetName() string {
	return "ProfileManagementState"
}
