package query

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	queryports "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
)

type NotificationQueryService struct {
	db *sql.DB
}

func NewNotificationQueryService(db *sql.DB) *NotificationQueryService {
	return &NotificationQueryService{db: db}
}

func (q *NotificationQueryService) FindAllActive(ctx context.Context, at time.Time) ([]queryports.ActiveDuty, error) {
	const query = `
		SELECT id, group_id, team_id, start_date, end_date
		FROM duty
		WHERE group_id IS NOT NULL
		  AND start_date <= ?
		  AND end_date > ?
		ORDER BY start_date, id
	`

	rows, err := q.db.QueryContext(ctx, query, at, at)
	if err != nil {
		return nil, fmt.Errorf("find active duties: %w", err)
	}
	defer rows.Close()

	result := make([]queryports.ActiveDuty, 0)
	for rows.Next() {
		var idBytes, groupIDBytes, teamIDBytes []byte
		var item queryports.ActiveDuty

		if err := rows.Scan(&idBytes, &groupIDBytes, &teamIDBytes, &item.StartAt, &item.EndAt); err != nil {
			return nil, fmt.Errorf("scan active duty: %w", err)
		}

		item.ID, err = uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("decode active duty id: %w", err)
		}
		item.GroupID, err = uuid.FromBytes(groupIDBytes)
		if err != nil {
			return nil, fmt.Errorf("decode active duty group id: %w", err)
		}
		item.TeamID, err = uuid.FromBytes(teamIDBytes)
		if err != nil {
			return nil, fmt.Errorf("decode active duty team id: %w", err)
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active duties: %w", err)
	}

	return result, nil
}

func (q *NotificationQueryService) FindUserIDsByTeamID(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT DISTINCT user_id
		FROM (
			SELECT u.id AS user_id
			FROM user u
			WHERE u.team_id = ?
			  AND u.deleted_at IS NULL

			UNION

			SELECT t.leader_id AS user_id
			FROM team t
			JOIN user u ON u.id = t.leader_id
			WHERE t.id = ?
			  AND t.leader_id IS NOT NULL
			  AND u.deleted_at IS NULL
		) AS team_users
		WHERE user_id IS NOT NULL
	`

	teamIDBytes, err := teamID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal team id: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, query, teamIDBytes, teamIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find team member ids: %w", err)
	}
	defer rows.Close()

	return scanUUIDRows(rows, "scan team member id")
}

func (q *NotificationQueryService) FindActiveParticipantIDsByDutyID(ctx context.Context, dutyID uuid.UUID) ([]uuid.UUID, error) {
	dutyIDBytes, err := dutyID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal duty id: %w", err)
	}
	rows, err := q.db.QueryContext(ctx, `SELECT p.participant_id FROM duty_participants p JOIN user u ON u.id = p.participant_id WHERE p.duty_id = ? AND p.excluded_at IS NULL AND u.deleted_at IS NULL`, dutyIDBytes)
	if err != nil {
		return nil, fmt.Errorf("find active duty participant ids: %w", err)
	}
	defer rows.Close()
	return scanUUIDRows(rows, "scan duty participant id")
}

func (q *NotificationQueryService) GetByDutyID(ctx context.Context, dutyID uuid.UUID, teamID uuid.UUID) (queryports.DutyFinishReminderState, error) {
	dutyIDBytes, err := dutyID.MarshalBinary()
	if err != nil {
		return queryports.DutyFinishReminderState{}, fmt.Errorf("marshal duty id: %w", err)
	}
	teamIDBytes, err := teamID.MarshalBinary()
	if err != nil {
		return queryports.DutyFinishReminderState{}, fmt.Errorf("marshal team id: %w", err)
	}

	freeTaskCount, err := q.getFreeTaskCount(ctx, dutyIDBytes)
	if err != nil {
		return queryports.DutyFinishReminderState{}, err
	}

	memberIDs, err := q.FindActiveParticipantIDsByDutyID(ctx, dutyID)
	if err != nil {
		return queryports.DutyFinishReminderState{}, err
	}

	usersBelowAssignedGoalIDs, err := q.getUsersBelowAssignedGoal(ctx, dutyIDBytes, teamIDBytes, memberIDs)
	if err != nil {
		return queryports.DutyFinishReminderState{}, err
	}

	userIDs, err := q.getUsersWithIncompleteTasks(ctx, dutyIDBytes)
	if err != nil {
		return queryports.DutyFinishReminderState{}, err
	}

	return queryports.DutyFinishReminderState{
		FreeTaskCount:              freeTaskCount,
		UsersBelowAssignedGoalIDs:  usersBelowAssignedGoalIDs,
		UsersWithIncompleteTaskIDs: userIDs,
	}, nil
}

func (q *NotificationQueryService) getFreeTaskCount(ctx context.Context, dutyID []byte) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM duty_task
		WHERE duty_id = ?
		  AND assignee_id IS NULL
	`

	var count int
	if err := q.db.QueryRowContext(ctx, query, dutyID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count free duty tasks: %w", err)
	}

	return count, nil
}

func (q *NotificationQueryService) getUsersWithIncompleteTasks(ctx context.Context, dutyID []byte) ([]uuid.UUID, error) {
	const query = `
		SELECT DISTINCT assignee_id
		FROM duty_task
		WHERE duty_id = ?
		  AND assignee_id IS NOT NULL
		  AND completion_date IS NULL
	`

	rows, err := q.db.QueryContext(ctx, query, dutyID)
	if err != nil {
		return nil, fmt.Errorf("find users with incomplete tasks: %w", err)
	}
	defer rows.Close()

	return scanUUIDRows(rows, "scan user with incomplete task")
}

func (q *NotificationQueryService) getUsersBelowAssignedGoal(
	ctx context.Context,
	dutyID []byte,
	teamID []byte,
	memberIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	memberCount := len(memberIDs)
	if memberCount == 0 {
		return nil, nil
	}

	requiredGoal, err := q.getAssignedGoalPerMember(ctx, dutyID, teamID, memberCount)
	if err != nil {
		return nil, err
	}
	if requiredGoal <= 0 {
		return nil, nil
	}

	const query = `
		SELECT user_id
		FROM (
			SELECT team_users.user_id AS user_id, COALESCE(SUM(task.cost), 0) AS assigned_cost
			FROM (
			SELECT participant_id AS user_id
			FROM duty_participants
			WHERE duty_id = ? AND excluded_at IS NULL
			) AS team_users
			LEFT JOIN duty_task dt
				ON dt.duty_id = ?
			   AND dt.assignee_id = team_users.user_id
			LEFT JOIN task
				ON task.id = dt.task_id
			GROUP BY team_users.user_id
		) AS user_costs
		WHERE assigned_cost < ?
	`

	rows, err := q.db.QueryContext(ctx, query, dutyID, dutyID, requiredGoal)
	if err != nil {
		return nil, fmt.Errorf("find users below assigned goal: %w", err)
	}
	defer rows.Close()

	return scanUUIDRows(rows, "scan user below assigned goal")
}

func (q *NotificationQueryService) getAssignedGoalPerMember(
	ctx context.Context,
	dutyID []byte,
	teamID []byte,
	memberCount int,
) (int, error) {
	const query = `
		SELECT COALESCE(SUM(task.cost), 0)
		FROM duty_task dt
		JOIN task ON task.id = dt.task_id
		WHERE dt.duty_id = ?
	`

	var totalCost int
	if err := q.db.QueryRowContext(ctx, query, dutyID).Scan(&totalCost); err != nil {
		return 0, fmt.Errorf("count total duty cost: %w", err)
	}

	if memberCount == 0 {
		return 0, nil
	}

	return (totalCost + memberCount - 1) / memberCount, nil
}

func scanUUIDRows(rows *sql.Rows, action string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	for rows.Next() {
		var idBytes []byte
		if err := rows.Scan(&idBytes); err != nil {
			return nil, fmt.Errorf("%s: %w", action, err)
		}

		id, err := uuid.FromBytes(idBytes)
		if err != nil {
			return nil, fmt.Errorf("%s decode uuid: %w", action, err)
		}

		result = append(result, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s iterate rows: %w", action, err)
	}

	return result, nil
}
