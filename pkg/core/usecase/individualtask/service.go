package individualtask

import (
	"context"
	"strings"
	"time"

	"dorm/pkg/core/domain/catalog"
	individual "dorm/pkg/core/domain/individualtask"
	"dorm/pkg/core/domain/structure"
	"dorm/pkg/core/domain/user"
	"dorm/pkg/core/ports/dto"
	query "dorm/pkg/core/ports/query"

	"github.com/google/uuid"
)

type Service struct {
	repo     individual.Repository
	users    user.Repository
	groups   structure.GroupRepository
	dorms    structure.DormitoryRepository
	areas    catalog.AreaRepository
	query    query.IndividualTaskQueryService
	location *time.Location
	now      func() time.Time
}

func NewService(r individual.Repository, u user.Repository, g structure.GroupRepository, d structure.DormitoryRepository, a catalog.AreaRepository, q query.IndividualTaskQueryService, l *time.Location) *Service {
	return &Service{r, u, g, d, a, q, l, time.Now}
}
func (s *Service) Create(c context.Context, actor uuid.UUID, req dto.IndividualTaskRequest) (*dto.IndividualTaskItem, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return nil, e
	}
	rid, e := uuid.Parse(strings.TrimSpace(req.ResidentID))
	if e != nil {
		return nil, individual.ErrInvalidResident
	}
	resident, e := s.users.FindByID(c, rid)
	if e != nil {
		return nil, e
	}
	if resident == nil {
		return nil, individual.ErrResidentNotFound
	}
	if resident.DormitoryID() == nil || !in(*resident.DormitoryID(), scope) {
		return nil, individual.ErrAccessDenied
	}
	dl, e := s.deadline(req.Deadline)
	if e != nil {
		return nil, e
	}
	if e = s.validateArea(c, req.AreaID, *resident.DormitoryID()); e != nil {
		return nil, e
	}
	t, e := individual.New(*resident.DormitoryID(), rid, req.AreaID, req.Title, req.RedemptionWeight, dl, s.now().In(s.location))
	if e != nil {
		return nil, e
	}
	if e = s.repo.Create(c, t); e != nil {
		return nil, e
	}
	return s.item(c, t.ID, actor, scope)
}
func (s *Service) Update(c context.Context, actor, id uuid.UUID, req dto.IndividualTaskRequest) (*dto.IndividualTaskItem, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return nil, e
	}
	t, e := s.repo.FindByID(c, id, false)
	if e != nil {
		return nil, e
	}
	if !in(t.DormitoryID, scope) {
		return nil, individual.ErrAccessDenied
	}
	if !t.CanEdit() {
		return nil, individual.ErrInvalidTransition
	}
	rid, e := uuid.Parse(strings.TrimSpace(req.ResidentID))
	if e != nil {
		return nil, individual.ErrInvalidResident
	}
	if rid != t.ResidentID {
		u, e := s.users.FindByID(c, rid)
		if e != nil {
			return nil, e
		}
		if u == nil {
			return nil, individual.ErrResidentNotFound
		}
		if u.DormitoryID() == nil || *u.DormitoryID() != t.DormitoryID {
			return nil, individual.ErrAccessDenied
		}
	}
	title, e := individual.NormalizeTitle(req.Title)
	if e != nil {
		return nil, e
	}
	w, e := individual.NormalizeWeight(req.RedemptionWeight)
	if e != nil {
		return nil, e
	}
	dl, e := s.deadline(req.Deadline)
	if e != nil {
		return nil, e
	}
	if e = s.validateArea(c, req.AreaID, t.DormitoryID); e != nil {
		return nil, e
	}
	t.ResidentID = rid
	t.AreaID = req.AreaID
	t.Title = title
	t.RedemptionWeight = w
	t.Deadline = dl
	t.UpdatedAt = s.now().In(s.location)
	if e = s.repo.Update(c, t, req.Version); e != nil {
		return nil, e
	}
	return s.item(c, id, actor, scope)
}
func (s *Service) Delete(c context.Context, actor, id uuid.UUID) error {
	scope, e := s.scope(c, actor)
	if e != nil {
		return e
	}
	t, e := s.repo.FindByID(c, id, true)
	if e != nil {
		return e
	}
	if !in(t.DormitoryID, scope) {
		return individual.ErrAccessDenied
	}
	return s.repo.Delete(c, id)
}
func (s *Service) Complete(c context.Context, actor, id uuid.UUID) (*dto.IndividualTaskItem, error) {
	t, e := s.repo.Complete(c, id, actor)
	if e != nil {
		return nil, e
	}
	return s.item(c, t.ID, actor, nil)
}
func (s *Service) Open(c context.Context, actor, id uuid.UUID) (*dto.IndividualTaskItem, error) {
	t, e := s.repo.Open(c, id, actor)
	if e != nil {
		return nil, e
	}
	return s.item(c, t.ID, actor, nil)
}
func (s *Service) Reject(c context.Context, actor, id uuid.UUID) (*dto.IndividualTaskItem, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return nil, e
	}
	t, e := s.repo.FindByID(c, id, false)
	if e != nil {
		return nil, e
	}
	if !in(t.DormitoryID, scope) {
		return nil, individual.ErrAccessDenied
	}
	t, e = s.repo.Reject(c, id)
	if e != nil {
		return nil, e
	}
	return s.item(c, t.ID, actor, scope)
}
func (s *Service) Verify(c context.Context, actor, id uuid.UUID) (*dto.IndividualTaskItem, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return nil, e
	}
	t, e := s.repo.FindByID(c, id, false)
	if e != nil {
		return nil, e
	}
	if !in(t.DormitoryID, scope) {
		return nil, individual.ErrAccessDenied
	}
	t, e = s.repo.Verify(c, id)
	if e != nil {
		return nil, e
	}
	return s.item(c, t.ID, actor, scope)
}
func (s *Service) Get(c context.Context, actor, id uuid.UUID) (*dto.IndividualTaskItem, error) {
	t, e := s.repo.FindByID(c, id, false)
	if e != nil {
		return nil, e
	}
	scope, _ := s.scope(c, actor)
	if t.ResidentID != actor && !in(t.DormitoryID, scope) {
		return nil, individual.ErrAccessDenied
	}
	return s.item(c, id, actor, scope)
}
func (s *Service) ListMine(c context.Context, actor uuid.UUID) (dto.IndividualTaskListResponse, error) {
	x, e := s.query.ListMine(c, actor)
	if e != nil {
		return dto.IndividualTaskListResponse{}, e
	}
	for i := range x {
		x[i].CanComplete = x[i].Status == "issued"
		x[i].CanOpen = x[i].Status == "completed"
	}
	return dto.IndividualTaskListResponse{Tasks: x}, nil
}
func (s *Service) ListResident(c context.Context, actor, resident uuid.UUID) (dto.IndividualTaskListResponse, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return dto.IndividualTaskListResponse{}, e
	}
	u, e := s.users.FindByID(c, resident)
	if e != nil {
		return dto.IndividualTaskListResponse{}, e
	}
	if u == nil {
		return dto.IndividualTaskListResponse{}, individual.ErrResidentNotFound
	}
	x, e := s.query.ListResident(c, resident, scope)
	if e != nil {
		return dto.IndividualTaskListResponse{}, e
	}
	return dto.IndividualTaskListResponse{Tasks: s.permissions(x, actor, scope)}, nil
}
func (s *Service) ListReview(c context.Context, actor uuid.UUID) (dto.IndividualTaskListResponse, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return dto.IndividualTaskListResponse{}, e
	}
	x, e := s.query.ListReview(c, scope)
	return dto.IndividualTaskListResponse{Tasks: s.permissions(x, actor, scope)}, e
}
func (s *Service) SearchResidents(c context.Context, actor uuid.UUID, q string) (dto.IndividualTaskResidentsResponse, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return dto.IndividualTaskResidentsResponse{}, e
	}
	x, e := s.query.SearchResidents(c, scope, q)
	return dto.IndividualTaskResidentsResponse{Residents: x}, e
}
func (s *Service) ListAreas(c context.Context, actor uuid.UUID, dorm int64) (dto.IndividualTaskAreasResponse, error) {
	scope, e := s.scope(c, actor)
	if e != nil {
		return dto.IndividualTaskAreasResponse{}, e
	}
	if !in(dorm, scope) {
		return dto.IndividualTaskAreasResponse{}, individual.ErrAccessDenied
	}
	x, e := s.query.ListAreas(c, dorm)
	return dto.IndividualTaskAreasResponse{Areas: x}, e
}
func (s *Service) scope(c context.Context, actor uuid.UUID) ([]int64, error) {
	u, e := s.users.FindByID(c, actor)
	if e != nil {
		return nil, e
	}
	if u == nil {
		return nil, individual.ErrAccessDenied
	}
	ds, e := s.dorms.FindAll(c)
	if e != nil {
		return nil, e
	}
	out := []int64{}
	for _, d := range ds {
		if d.LeaderID() != nil && *d.LeaderID() == actor {
			out = append(out, d.ID())
		}
	}
	if u.DormitoryID() != nil {
		gs, e := s.groups.FindByDormitoryID(c, *u.DormitoryID())
		if e != nil {
			return nil, e
		}
		for _, g := range gs {
			if g.LeaderID() != nil && *g.LeaderID() == actor && !in(*u.DormitoryID(), out) {
				out = append(out, *u.DormitoryID())
			}
		}
	}
	if len(out) > 0 {
		return out, nil
	}
	return nil, individual.ErrAccessDenied
}
func (s *Service) validateArea(c context.Context, id *int, dorm int64) error {
	if id == nil {
		return nil
	}
	a, e := s.areas.FindByID(c, *id)
	if e != nil {
		return e
	}
	if a == nil {
		return individual.ErrInvalidArea
	}
	if a.GroupID() == nil {
		return nil
	}
	g, e := s.groups.FindByID(c, *a.GroupID())
	if e != nil {
		return e
	}
	if g == nil || g.DormitoryID() != dorm {
		return individual.ErrInvalidArea
	}
	return nil
}
func (s *Service) deadline(raw *string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	v, e := time.ParseInLocation("2006-01-02", *raw, s.location)
	if e != nil {
		return nil, individual.ErrInvalidDeadline
	}
	today := s.now().In(s.location)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, s.location)
	if v.Before(today) {
		return nil, individual.ErrInvalidDeadline
	}
	return &v, nil
}
func (s *Service) item(c context.Context, id, actor uuid.UUID, scope []int64) (*dto.IndividualTaskItem, error) {
	x, e := s.query.Get(c, id)
	if e != nil {
		return nil, e
	}
	if x == nil {
		return nil, individual.ErrNotFound
	}
	*x = s.permissions([]dto.IndividualTaskItem{*x}, actor, scope)[0]
	return x, nil
}
func (s *Service) permissions(xs []dto.IndividualTaskItem, actor uuid.UUID, scope []int64) []dto.IndividualTaskItem {
	for i := range xs {
		m := in(xs[i].DormitoryID, scope)
		xs[i].CanEdit = m && xs[i].Status == "issued"
		xs[i].CanDelete = m && xs[i].Status != "verified"
		xs[i].CanVerify = m && xs[i].Status == "completed"
		xs[i].CanReject = xs[i].CanVerify
		xs[i].CanComplete = xs[i].Resident.ID == actor.String() && xs[i].Status == "issued"
		xs[i].CanOpen = xs[i].Resident.ID == actor.String() && xs[i].Status == "completed"
	}
	return xs
}
func in(id int64, x []int64) bool {
	for _, v := range x {
		if v == id {
			return true
		}
	}
	return false
}
