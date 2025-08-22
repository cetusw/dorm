package bot

type PaymentManagementState struct {
	baseState
}

// TODO: здесь можно реализовать логику для оплаты проживания через интерфейс телеграма

func (s *PaymentManagementState) GetName() string {
	return "PaymentManagementState"
}
