package dto

type ResidentListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RoomNumber string `json:"room_number"`
}

type ResidentListResponse struct {
	Users []ResidentListItem `json:"users"`
}

type ResidentDetails struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	Login      string  `json:"login"`
	Floor      *int    `json:"floor"`
	RoomNumber *string `json:"room_number"`
}

type CreateResidentRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	MiddleName  *string `json:"middle_name"`
	Login       string  `json:"login"`
	Password    string  `json:"password"`
	DormitoryID int64   `json:"dormitory_id"`
	Floor       *int    `json:"floor"`
	RoomNumber  *string `json:"room_number"`
}

type UpdateResidentRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	MiddleName  *string `json:"middle_name"`
	Login       string  `json:"login"`
	Password    string  `json:"password"`
	DormitoryID int64   `json:"dormitory_id"`
	Floor       *int    `json:"floor"`
	RoomNumber  *string `json:"room_number"`
}

type ResidentLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CurrentUserResponse struct {
	ID                   string `json:"id"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	CanManageDormitories bool   `json:"can_manage_dormitories"`
}
