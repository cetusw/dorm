package dutyparticipants

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	"github.com/google/uuid"
)

var (
	ErrAccessDenied          = errors.New("duty participants access denied")
	ErrDutyNotFound          = errors.New("duty not found")
	ErrNotCurrent            = errors.New("duty participants can only be edited for current duty")
	ErrCandidateOutsideGroup = errors.New("participant must belong to the duty dormitory")
)

type Service struct {
	users        user.Repository
	teams        structure.TeamRepository
	groups       structure.GroupRepository
	dormitories  structure.DormitoryRepository
	duties       duty.DutyRepository
	participants duty.ParticipantRepository
	now          func() time.Time
	location     *time.Location
}

func NewService(users user.Repository, teams structure.TeamRepository, groups structure.GroupRepository, dormitories structure.DormitoryRepository, duties duty.DutyRepository, participants duty.ParticipantRepository) *Service {
	return &Service{users: users, teams: teams, groups: groups, dormitories: dormitories, duties: duties, participants: participants, now: time.Now, location: time.Local}
}
func (s *Service) SetLocation(location *time.Location) {
	if location != nil {
		s.location = location
	}
}
func (s *Service) GetParticipants(ctx context.Context, actorID, dutyID uuid.UUID) ([]dto.DutyParticipantResponse, error) {
	d, _, err := s.requireManageable(ctx, actorID, dutyID)
	if err != nil {
		return nil, err
	}
	participants, err := s.participants.List(ctx, d.ID(), true)
	if err != nil {
		return nil, fmt.Errorf("list duty participants: %w", err)
	}
	result := make([]dto.DutyParticipantResponse, 0, len(participants))
	for _, p := range participants {
		u, err := s.users.FindByID(ctx, p.ParticipantID)
		if err != nil || u == nil {
			continue
		}
		item := dto.DutyParticipantResponse{ParticipantID: p.ParticipantID.String(), FullName: fullName(u), Type: string(p.Type), IsLeader: d.LeaderID() != nil && *d.LeaderID() == p.ParticipantID}
		if p.ExcludedAt != nil {
			v := p.ExcludedAt.In(s.location).Format(time.RFC3339)
			item.ExcludedAt = &v
		}
		if u.TeamID() != nil {
			id := u.TeamID().String()
			item.TeamID = &id
			if t, _ := s.teams.FindByID(ctx, *u.TeamID()); t != nil {
				if leader, _ := s.users.FindByID(ctx, t.LeaderID()); leader != nil {
					name := fullName(leader)
					item.TeamName = &name
				}
			}
		}
		result = append(result, item)
	}
	return result, nil
}
func (s *Service) GetCandidates(ctx context.Context, actorID, dutyID uuid.UUID) ([]dto.DutyParticipantCandidateResponse, error) {
	_, g, err := s.requireManageable(ctx, actorID, dutyID)
	if err != nil {
		return nil, err
	}
	// Residents are currently scoped to the dormitory. A resident may not yet
	// have a permanent Team, so deriving candidates through user.team_id would
	// incorrectly hide them from the duty roster.
	users, err := s.users.FindByDormitoryID(ctx, g.DormitoryID())
	if err != nil {
		return nil, err
	}
	out := []dto.DutyParticipantCandidateResponse{}
	for _, u := range users {
		item := dto.DutyParticipantCandidateResponse{ParticipantID: u.ID().String(), FullName: fullName(u)}
		if u.TeamID() != nil {
			if t, findErr := s.teams.FindByID(ctx, *u.TeamID()); findErr == nil && t != nil {
				id := u.TeamID().String()
				item.TeamID = &id
				if leader, leaderErr := s.users.FindByID(ctx, t.LeaderID()); leaderErr == nil && leader != nil {
					name := fullName(leader)
					item.TeamName = &name
				}
			}
		}
		out = append(out, item)
	}
	// FindByDormitoryID is ordered by name in MySQL; retain an explicit stable
	// ordering for other repository implementations too.
	sort.Slice(out, func(i, j int) bool { return out[i].FullName < out[j].FullName })
	return out, nil
}
func (s *Service) AddParticipant(ctx context.Context, actorID, dutyID, participantID uuid.UUID) error {
	d, g, err := s.requireManageable(ctx, actorID, dutyID)
	if err != nil {
		return err
	}
	if err := s.ensureSameDormitory(ctx, participantID, g.DormitoryID()); err != nil {
		return err
	}
	return s.participants.Add(ctx, duty.DutyParticipant{DutyID: dutyID, ParticipantID: participantID, Type: duty.ParticipantTypeTemporary}, d.Start(), d.End())
}
func (s *Service) ExcludeParticipant(ctx context.Context, actorID, dutyID, participantID uuid.UUID, replacement *uuid.UUID) error {
	d, _, err := s.requireManageable(ctx, actorID, dutyID)
	if err != nil {
		return err
	}
	return s.participants.Exclude(ctx, d.ID(), participantID, replacement, s.currentTime())
}
func (s *Service) RestoreParticipant(ctx context.Context, actorID, dutyID, participantID uuid.UUID) error {
	d, _, err := s.requireManageable(ctx, actorID, dutyID)
	if err != nil {
		return err
	}
	return s.participants.Restore(ctx, d.ID(), participantID, d.Start(), d.End())
}
func (s *Service) ChangeLeader(ctx context.Context, actorID, dutyID, leaderID uuid.UUID) error {
	d, _, err := s.requireManageable(ctx, actorID, dutyID)
	if err != nil {
		return err
	}
	return s.participants.ChangeLeader(ctx, d.ID(), leaderID)
}
func (s *Service) requireManageable(ctx context.Context, actorID, dutyID uuid.UUID) (*duty.Duty, *structure.Group, error) {
	d, err := s.duties.FindByID(ctx, dutyID)
	if err != nil {
		return nil, nil, err
	}
	if d == nil {
		return nil, nil, ErrDutyNotFound
	}
	if !d.IsActiveAt(s.currentTime()) {
		return nil, nil, ErrNotCurrent
	}
	team, err := s.teams.FindByID(ctx, d.TeamID())
	if err != nil || team == nil {
		return nil, nil, ErrDutyNotFound
	}
	group, err := s.groups.FindByID(ctx, team.GroupID())
	if err != nil || group == nil {
		return nil, nil, ErrDutyNotFound
	}
	if d.LeaderID() != nil && *d.LeaderID() == actorID {
		return d, group, nil
	}
	if group.LeaderID() != nil && *group.LeaderID() == actorID {
		return d, group, nil
	}
	dorm, err := s.dormitories.FindByID(ctx, group.DormitoryID())
	if err != nil {
		return nil, nil, err
	}
	if dorm != nil && dorm.LeaderID() != nil && *dorm.LeaderID() == actorID {
		return d, group, nil
	}
	return nil, nil, ErrAccessDenied
}
func (s *Service) ensureSameDormitory(ctx context.Context, userID uuid.UUID, dormitoryID int64) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil || u.DormitoryID() == nil || *u.DormitoryID() != dormitoryID {
		return ErrCandidateOutsideGroup
	}
	return nil
}
func (s *Service) currentTime() time.Time {
	now := s.now()
	if s.location != nil {
		return now.In(s.location)
	}
	return now
}
func fullName(u *user.User) string {
	if u.MiddleName() != nil && *u.MiddleName() != "" {
		return u.LastName() + " " + u.FirstName() + " " + *u.MiddleName()
	}
	return u.LastName() + " " + u.FirstName()
}
