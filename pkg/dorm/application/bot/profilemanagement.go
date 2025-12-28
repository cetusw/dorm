package bot

type ProfileInfoState struct {
	baseState
}

func (s *ProfileInfoState) GetName() string {
	return "ProfileInfoState"
}
