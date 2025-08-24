package bot

type MainState struct {
	baseState
}

func (s *MainState) GetName() string {
	return "MainState"
}
