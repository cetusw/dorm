package bot

type TeamManagementState struct {
	baseState
}

// TODO: реализовать логику для управления командой

func (s *TeamManagementState) GetName() string {
	return "TeamManagementState"
}
