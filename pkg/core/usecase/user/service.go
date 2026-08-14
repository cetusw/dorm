package user

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"dorm/pkg/core/ports/dto"
	ports "dorm/pkg/core/ports/query"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
)

var (
	cyrillicNamePattern   = regexp.MustCompile(`^[\p{Cyrillic}\s-]+$`)
	residentLoginPattern  = regexp.MustCompile(`^[a-z0-9._-]+$`)
	residentPasswordRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)
)

type Service struct {
	userRepo      user.Repository
	dormitoryRepo structure.DormitoryRepository
	groupRepo     structure.GroupRepository
	queryService  ports.UserQueryService
}

func NewUserService(
	userRepo user.Repository,
	dormitoryRepo structure.DormitoryRepository,
	groupRepo structure.GroupRepository,
	queryService ports.UserQueryService,
) *Service {
	return &Service{
		userRepo:      userRepo,
		dormitoryRepo: dormitoryRepo,
		groupRepo:     groupRepo,
		queryService:  queryService,
	}
}

func (s *Service) AuthenticateResident(ctx context.Context, login string, password string) (*user.User, error) {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	if login == "" || password == "" {
		return nil, fmt.Errorf("логин и пароль обязательны")
	}

	u, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("find user by login: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("неверный логин или пароль")
	}

	passwordHash := hashResidentPassword(password)
	if subtle.ConstantTimeCompare([]byte(u.PasswordHash()), []byte(passwordHash)) != 1 {
		return nil, fmt.Errorf("неверный логин или пароль")
	}

	return u, nil
}

func hashResidentPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func hashRequiredResidentPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", fmt.Errorf("пароль обязателен")
	}

	return hashResidentPassword(password), nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.CurrentUserResponse, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find current user: %w", err)
	}
	if u == nil {
		return nil, nil
	}

	canManageDormitories, err := s.dormitoryRepo.ExistsByLeaderID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check dormitory management access: %w", err)
	}

	canManagePenalties, err := s.canManagePenalties(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("check penalty management access: %w", err)
	}

	canManageWarehouse, err := s.canManageWarehouse(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("check warehouse management access: %w", err)
	}

	return &dto.CurrentUserResponse{
		ID:                   u.ID().String(),
		FirstName:            u.FirstName(),
		LastName:             u.LastName(),
		CanManageDormitories: canManageDormitories,
		CanManagePenalties:   canManagePenalties,
		CanManageWarehouse:   canManageWarehouse,
	}, nil
}

func (s *Service) canManagePenalties(ctx context.Context, currentUser *user.User) (bool, error) {
	canManageDormitories, err := s.dormitoryRepo.ExistsByLeaderID(ctx, currentUser.ID())
	if err != nil {
		return false, fmt.Errorf("check dormitory leadership for penalties: %w", err)
	}
	if canManageDormitories {
		return true, nil
	}

	if currentUser.DormitoryID() == nil {
		return false, nil
	}

	groups, err := s.groupRepo.FindByDormitoryID(ctx, *currentUser.DormitoryID())
	if err != nil {
		return false, fmt.Errorf("load dormitory groups for penalties: %w", err)
	}

	for _, group := range groups {
		if group.LeaderID() != nil && *group.LeaderID() == currentUser.ID() {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) canManageWarehouse(ctx context.Context, currentUser *user.User) (bool, error) {
	if currentUser.DormitoryID() == nil {
		return false, nil
	}

	dormitory, err := s.dormitoryRepo.FindByID(ctx, *currentUser.DormitoryID())
	if err != nil {
		return false, fmt.Errorf("load current dormitory for warehouse: %w", err)
	}
	if dormitory != nil && dormitory.LeaderID() != nil && *dormitory.LeaderID() == currentUser.ID() {
		return true, nil
	}

	groups, err := s.groupRepo.FindByDormitoryID(ctx, *currentUser.DormitoryID())
	if err != nil {
		return false, fmt.Errorf("load dormitory groups for warehouse: %w", err)
	}

	for _, group := range groups {
		if group.LeaderID() != nil && *group.LeaderID() == currentUser.ID() {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) GetResidentsResponse(ctx context.Context, dormitoryID int64) (dto.ResidentListResponse, error) {
	items, err := s.queryService.GetResidentsByDormitoryID(ctx, dormitoryID)
	if err != nil {
		return dto.ResidentListResponse{}, fmt.Errorf("load residents by dormitory: %w", err)
	}

	return dto.ResidentListResponse{
		Users: items,
	}, nil
}

func (s *Service) GetResidentDetails(ctx context.Context, id uuid.UUID) (*dto.ResidentDetails, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find resident: %w", err)
	}
	if u == nil {
		return nil, nil
	}

	return residentDetailsFromUser(u), nil
}

func (s *Service) CreateResident(ctx context.Context, req dto.CreateResidentRequest) (*dto.ResidentDetails, error) {
	input, err := s.normalizeResidentInput(ctx, residentInput{
		firstName:   req.FirstName,
		lastName:    req.LastName,
		middleName:  req.MiddleName,
		login:       req.Login,
		password:    req.Password,
		dormitoryID: req.DormitoryID,
		floor:       req.Floor,
		roomNumber:  req.RoomNumber,
	})
	if err != nil {
		return nil, err
	}

	passwordHash, err := hashRequiredResidentPassword(input.password)
	if err != nil {
		return nil, err
	}

	resident, err := user.NewUser(input.firstName, input.lastName, input.login, passwordHash)
	if err != nil {
		return nil, err
	}
	if input.middleName != nil {
		resident.SetMiddleName(*input.middleName)
	}
	resident.MoveInto(input.dormitoryID, input.roomNumberValue())
	if input.floor != nil {
		resident.SetFloorNumber(*input.floor)
	}

	if err := s.userRepo.Save(ctx, resident); err != nil {
		return nil, fmt.Errorf("create resident: %w", err)
	}

	return residentDetailsFromUser(resident), nil
}

func (s *Service) UpdateResident(ctx context.Context, id uuid.UUID, req dto.UpdateResidentRequest) (*dto.ResidentDetails, error) {
	resident, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find resident: %w", err)
	}
	if resident == nil {
		return nil, nil
	}

	input, err := s.normalizeResidentInput(ctx, residentInput{
		firstName:   req.FirstName,
		lastName:    req.LastName,
		middleName:  req.MiddleName,
		login:       req.Login,
		password:    req.Password,
		dormitoryID: req.DormitoryID,
		floor:       req.Floor,
		roomNumber:  req.RoomNumber,
		excludeID:   &id,
	})
	if err != nil {
		return nil, err
	}

	if err := resident.Rename(input.firstName, input.lastName); err != nil {
		return nil, err
	}
	resident.SetMiddleName(input.middleNameValue())
	resident.MoveInto(input.dormitoryID, input.roomNumberValue())
	if input.floor != nil {
		resident.SetFloorNumber(*input.floor)
	} else {
		resident.SetFloorNumber(0)
	}
	if err := resident.SetLogin(input.login); err != nil {
		return nil, err
	}

	passwordHash, err := hashRequiredResidentPassword(input.password)
	if err != nil {
		return nil, err
	}
	if err := resident.SetPasswordHash(passwordHash); err != nil {
		return nil, err
	}

	if err := s.userRepo.Save(ctx, resident); err != nil {
		return nil, fmt.Errorf("update resident: %w", err)
	}

	return residentDetailsFromUser(resident), nil
}

func (s *Service) DeleteResident(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("delete resident: %w", err)
	}

	return nil
}

type residentInput struct {
	firstName   string
	lastName    string
	middleName  *string
	login       string
	password    string
	dormitoryID int64
	floor       *int
	roomNumber  *string
	excludeID   *uuid.UUID
}

func (i residentInput) middleNameValue() string {
	if i.middleName == nil {
		return ""
	}

	return *i.middleName
}

func (i residentInput) roomNumberValue() string {
	if i.roomNumber == nil {
		return ""
	}

	return *i.roomNumber
}

func (s *Service) normalizeResidentInput(ctx context.Context, input residentInput) (residentInput, error) {
	firstName, err := validateResidentName(input.firstName, "имя")
	if err != nil {
		return residentInput{}, err
	}

	lastName, err := validateResidentName(input.lastName, "фамилия")
	if err != nil {
		return residentInput{}, err
	}

	middleName, err := validateOptionalResidentName(input.middleName, "отчество")
	if err != nil {
		return residentInput{}, err
	}

	login, err := validateResidentLogin(input.login)
	if err != nil {
		return residentInput{}, err
	}

	password, err := validateResidentPassword(input.password)
	if err != nil {
		return residentInput{}, err
	}

	roomNumber, err := validateOptionalResidentText(input.roomNumber, "комната", 255)
	if err != nil {
		return residentInput{}, err
	}

	if input.floor != nil && (*input.floor < -1000 || *input.floor > 1000) {
		return residentInput{}, fmt.Errorf("этаж должен быть в допустимом диапазоне")
	}

	if input.dormitoryID <= 0 {
		return residentInput{}, fmt.Errorf("общежитие обязательно")
	}

	dormitory, err := s.dormitoryRepo.FindByID(ctx, input.dormitoryID)
	if err != nil {
		return residentInput{}, fmt.Errorf("find dormitory: %w", err)
	}
	if dormitory == nil {
		return residentInput{}, fmt.Errorf("общежитие не найдено")
	}

	existingUser, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		return residentInput{}, fmt.Errorf("find resident by login: %w", err)
	}
	if existingUser != nil && (input.excludeID == nil || existingUser.ID() != *input.excludeID) {
		return residentInput{}, fmt.Errorf("логин уже занят")
	}

	return residentInput{
		firstName:   firstName,
		lastName:    lastName,
		middleName:  middleName,
		login:       login,
		password:    password,
		dormitoryID: input.dormitoryID,
		floor:       input.floor,
		roomNumber:  roomNumber,
		excludeID:   input.excludeID,
	}, nil
}

func residentDetailsFromUser(u *user.User) *dto.ResidentDetails {
	if u == nil {
		return nil
	}

	return &dto.ResidentDetails{
		ID:         u.ID().String(),
		FirstName:  u.FirstName(),
		LastName:   u.LastName(),
		MiddleName: u.MiddleName(),
		Login:      u.Login(),
		Floor:      u.FloorNumber(),
		RoomNumber: u.RoomNumber(),
	}
}

func validateResidentName(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("поле \"%s\" обязательно", field)
	}
	if utf8.RuneCountInString(trimmed) > 255 {
		return "", fmt.Errorf("%s не должны превышать 255 символов", field)
	}
	if !cyrillicNamePattern.MatchString(trimmed) {
		return "", fmt.Errorf("%s должны содержать только кириллицу, пробел или дефис", field)
	}

	return trimmed, nil
}

func validateOptionalResidentName(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	validated, err := validateResidentName(trimmed, field)
	if err != nil {
		return nil, err
	}

	return &validated, nil
}

func validateResidentLogin(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("логин обязателен")
	}
	if utf8.RuneCountInString(trimmed) > 255 {
		return "", fmt.Errorf("логин не должен превышать 255 символов")
	}
	if !residentLoginPattern.MatchString(trimmed) {
		return "", fmt.Errorf("логин должен содержать только латинские буквы, цифры, точку, дефис или подчеркивание")
	}

	return trimmed, nil
}

func validateResidentPassword(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("пароль обязателен")
	}
	if utf8.RuneCountInString(trimmed) < 6 {
		return "", fmt.Errorf("пароль должен содержать минимум 6 символов")
	}
	if !residentPasswordRegex.MatchString(trimmed) {
		return "", fmt.Errorf("пароль должен содержать только латинские буквы и цифры")
	}

	return trimmed, nil
}

func validateOptionalResidentText(value *string, field string, maxLength int) (*string, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > maxLength {
		return nil, fmt.Errorf("%s не должна превышать %d символов", field, maxLength)
	}

	return &trimmed, nil
}
