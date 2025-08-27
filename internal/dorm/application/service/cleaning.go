package service

import (
	"dorm/internal/common/utils"
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure/mysql/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CleaningService struct {
	sheetsService                *SheetsService
	userService                  *UserService
	dutyService                  *DutyService
	dutyTaskService              *DutyTaskService
	teamService                  *TeamService
	taskService                  *TaskService
	specialTaskAssignmentService *SpecialTaskAssignmentService
}

func NewCleaningService(
	sheetsService *SheetsService,
	userService *UserService,
	dutyService *DutyService,
	dutyTaskService *DutyTaskService,
	teamService *TeamService,
	taskService *TaskService,
	specialTaskAssignmentService *SpecialTaskAssignmentService,
) *CleaningService {
	return &CleaningService{
		sheetsService:                sheetsService,
		userService:                  userService,
		dutyService:                  dutyService,
		dutyTaskService:              dutyTaskService,
		teamService:                  teamService,
		taskService:                  taskService,
		specialTaskAssignmentService: specialTaskAssignmentService,
	}
}

func (s *CleaningService) StartNewWeek() error {
	startTime := utils.NowMoscow()
	endTime := startTime.AddDate(0, 0, 6)
	sheetTitle := fmt.Sprintf("%s-%s", startTime.Format("02.01"), endTime.Format("02.01"))

	newDutyTeamID, err := s.getNewDutyTeamID()
	if err != nil {
		return err
	}
	var duty *model.Duty
	duty, err = s.dutyService.CreateNewDuty(newDutyTeamID, startTime, endTime)
	if err != nil {
		return err
	}
	team, err := s.teamService.GetTeam(duty.TeamID)
	if err != nil {
		return err
	}
	taskIDs, err := s.taskService.GetAllTaskIDsByFrequency(7)
	if err != nil {
		return err
	}
	err = s.dutyTaskService.SetDutyTasks(duty.DutyID, taskIDs, team.TeamLeaderID)
	if err != nil {
		return err
	}
	var teamColor string
	teamColor, err = s.teamService.GetTeamColor(newDutyTeamID)
	if err != nil {
		return err
	}
	var dutyTasksReadable []model.DutyTaskView
	dutyTasksReadable, err = s.dutyTaskService.GetDutyTasksReadable(duty.DutyID)
	if err != nil {
		return err
	}
	users, err := s.userService.GetAllUsers()
	if err != nil {
		return err
	}
	err = s.sheetsService.CreateWeeklySheet(sheetTitle, teamColor, newDutyTeamID, dutyTasksReadable, users)
	if err != nil {
		return err
	}
	return nil
}

func (s *CleaningService) AssignSpecialTasks(frequency int) error {
	taskIDs, users, err := s.loadTasksAndUsers(frequency)
	if err != nil {
		return err
	}
	if len(taskIDs) == 0 {
		return nil
	}
	now := utils.NowMoscow()
	slots := weekSlots(now, frequency)

	startIdxByTask, err := s.initStartIdxByTask(users, taskIDs)
	if err != nil {
		return err
	}

	rows := s.assignForWeek(users, taskIDs, slots, startIdxByTask)

	if len(rows) > 0 {
		if err := s.specialTaskAssignmentService.StoreBatch(rows); err != nil {
			return fmt.Errorf("store assignments: %v", err)
		}
	}
	return nil
}

func (s *CleaningService) loadTasksAndUsers(frequency int) ([]uuid.UUID, []model.User, error) {
	taskIDs, err := s.taskService.GetAllTaskIDsByFrequency(frequency)
	if err != nil {
		return nil, nil, fmt.Errorf("get tasks by frequency: %v", err)
	}
	users, err := s.userService.GetAllUsers()
	if err != nil {
		return nil, nil, fmt.Errorf("load users: %v", err)
	}
	if len(users) == 0 {
		return nil, nil, fmt.Errorf("no users to assign")
	}
	return taskIDs, users, nil
}

func (s *CleaningService) initStartIdxByTask(users []model.User, taskIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	startIdxByTask := make(map[uuid.UUID]int, len(taskIDs))
	for _, taskID := range taskIDs {
		idx, err := s.computeStartIdx(users, taskID)
		if err != nil {
			return nil, err
		}
		startIdxByTask[taskID] = idx
	}
	return startIdxByTask, nil
}

func (s *CleaningService) assignForWeek(
	users []model.User,
	taskIDs []uuid.UUID,
	slots []time.Time,
	startIdxByTask map[uuid.UUID]int,
) []repository.SpecialTaskAssignmentRow {
	var rows []repository.SpecialTaskAssignmentRow
	for _, at := range slots {
		slotRows := s.assignForSlot(users, taskIDs, at, startIdxByTask)
		rows = append(rows, slotRows...)
	}

	return rows
}

// Распределяем задачи в одном временном слоте, обновляя указатели startIdxByTask
func (s *CleaningService) assignForSlot(
	users []model.User,
	taskIDs []uuid.UUID,
	at time.Time,
	startIdxByTask map[uuid.UUID]int,
) []repository.SpecialTaskAssignmentRow {
	assignedNow := make(map[uuid.UUID]struct{}, len(users))
	var rows []repository.SpecialTaskAssignmentRow

	for _, taskID := range taskIDs {
		startIdx := startIdxByTask[taskID]
		userID := pickNextAvailableUser(users, startIdx, assignedNow)

		assignedNow[userID] = struct{}{}

		rows = append(rows, repository.SpecialTaskAssignmentRow{
			AssignmentID:   uuid.New(),
			TaskID:         taskID,
			AssigneeID:     userID,
			AssignmentDate: at,
			CompletionDate: nil,
		})

		nextIdx := indexOfUser(users, userID)
		if nextIdx < 0 {
			nextIdx = 0
		}
		startIdxByTask[taskID] = (nextIdx + 1) % len(users)
	}
	return rows
}

func weekSlots(from time.Time, frequency int) []time.Time {
	if frequency <= 0 {
		frequency = 1
	}
	end := from.AddDate(0, 0, 7)
	step := time.Duration(frequency) * 24 * time.Hour

	var res []time.Time
	for t := from; !t.After(end); t = t.Add(step) {
		res = append(res, t)
	}
	return res
}

func (s *CleaningService) UpdateCurrentSheet() error {
	duty, err := s.dutyService.GetCurrentDuty()
	if err != nil {
		return err
	}
	sheetTitle := fmt.Sprintf("%s-%s", duty.Start.Format("02.01"), duty.End.Format("02.01"))

	dutyTasksReadable, err := s.dutyTaskService.GetDutyTasksReadable(duty.DutyID)
	if err != nil {
		return fmt.Errorf("failed to get readable duty tasks for sheet update: %w", err)
	}

	err = s.sheetsService.UpdateWeeklySheet(sheetTitle, dutyTasksReadable)
	if err != nil {
		return fmt.Errorf("failed to update weekly sheet: %w", err)
	}

	return nil
}

func (s *CleaningService) getNewDutyTeamID() (int, error) {
	teamIDs, err := s.teamService.GetSortedTeamIDs()
	if err != nil {
		return 0, err
	}
	var lastDutyTeamID int
	lastDutyTeamID, err = s.dutyService.GetLastDutyTeamID()
	if err != nil {
		return 0, err
	}

	nextTeamIndex := 0

	for i, id := range teamIDs {
		if id == lastDutyTeamID {
			nextTeamIndex = (i + 1) % len(teamIDs)
			break
		}
	}

	return teamIDs[nextTeamIndex], nil
}

// Берём следующего после последнего исполнителя из истории.
// Если истории нет или пользователь не найден среди текущих, начинаем с 0.
func (s *CleaningService) computeStartIdx(users []model.User, taskID uuid.UUID) (int, error) {
	last, err := s.specialTaskAssignmentService.FindLatestByTaskID(taskID)
	if err != nil {
		return 0, fmt.Errorf("latest assignment for task %s: %v", taskID, err)
	}
	if last == nil {
		return 0, nil
	}
	idx := indexOfUser(users, last.AssigneeID)
	if idx < 0 {
		return 0, nil
	}
	return (idx + 1) % len(users), nil
}

// Ищем первого не назначенного в этом запуске, обход по кругу.
// Если всех уже назначили (задач больше, чем пользователей) — берём startIdx.
func pickNextAvailableUser(users []model.User, startIdx int, assigned map[uuid.UUID]struct{}) uuid.UUID {
	for step := 0; step < len(users); step++ {
		u := users[(startIdx+step)%len(users)].UserID
		if _, used := assigned[u]; !used {
			return u
		}
	}
	return users[startIdx].UserID
}

// Готовим записи для истории.
func buildSpecialAssignmentRows(assignments map[uuid.UUID]uuid.UUID, at time.Time) []repository.SpecialTaskAssignmentRow {
	rows := make([]repository.SpecialTaskAssignmentRow, 0, len(assignments))
	for taskID, userID := range assignments {
		rows = append(rows, repository.SpecialTaskAssignmentRow{
			AssignmentID:   uuid.New(),
			TaskID:         taskID,
			AssigneeID:     userID,
			AssignmentDate: at,
			CompletionDate: nil,
		})
	}
	return rows
}

// Находим индекс пользователя по UUID. Если нет — -1.
func indexOfUser(users []model.User, id uuid.UUID) int {
	for i, u := range users {
		if u.UserID == id {
			return i
		}
	}
	return -1
}
