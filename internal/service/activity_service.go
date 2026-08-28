package service

import (
	"errors"
	"math"
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ActivityService interface {
	Start(uuid.UUID, dto.StartActivityDto) (*model.Activity, error)
	Get(uuid.UUID, uuid.UUID) (*model.Activity, error)
	List(uuid.UUID) ([]model.Activity, error)
	AddPoints(uuid.UUID, uuid.UUID, dto.AppendActivityPointsDto) (*model.Activity, error)
	SetStatus(uuid.UUID, uuid.UUID, model.ActivityStatus, time.Time) (*model.Activity, error)
	Cancel(uuid.UUID, uuid.UUID) error
}
type activityService struct {
	repo repository.ActivityRepository
	pets repository.PetRepository
}

func NewActivityService(r repository.ActivityRepository, p repository.PetRepository) ActivityService {
	return &activityService{r, p}
}
func (s *activityService) Start(uid uuid.UUID, in dto.StartActivityDto) (*model.Activity, error) {
	privacy := in.Privacy
	if privacy == "" {
		privacy = model.ActivityPrivacyPrivate
	}
	a := &model.Activity{UserID: uid, PetID: in.PetID, Type: in.Type, Privacy: privacy, Status: model.ActivityStatusActive, StartedAt: in.StartedAt}
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return s.repo.FindByID(a.ID)
}
func (s *activityService) owned(uid, id uuid.UUID) (*model.Activity, error) {
	a, e := s.repo.FindByID(id)
	if e != nil || a == nil {
		return nil, errors.New("activity not found")
	}
	if a.UserID != uid {
		return nil, errors.New("access denied")
	}
	return a, nil
}
func (s *activityService) Get(uid, id uuid.UUID) (*model.Activity, error) { return s.owned(uid, id) }
func (s *activityService) List(uid uuid.UUID) ([]model.Activity, error) {
	return s.repo.ListByUser(uid)
}
func (s *activityService) AddPoints(uid, id uuid.UUID, in dto.AppendActivityPointsDto) (*model.Activity, error) {
	a, e := s.owned(uid, id)
	if e != nil {
		return nil, e
	}
	if a.Status != model.ActivityStatusActive {
		return nil, errors.New("activity is not active")
	}
	ps := make([]model.ActivityPoint, 0, len(in.Points))
	lastLat, lastLng := 0.0, 0.0
	var lastRecordedAt time.Time
	if len(a.Points) > 0 {
		last := a.Points[len(a.Points)-1]
		lastLat, lastLng = last.Lat, last.Lng
		lastRecordedAt = last.RecordedAt
	}
	for _, p := range in.Points {
		if p.AccuracyM != nil && *p.AccuracyM > 60 {
			continue
		}
		if p.RecordedAt.Before(a.StartedAt) || (!lastRecordedAt.IsZero() && !p.RecordedAt.After(lastRecordedAt)) {
			continue
		}
		if lastLat != 0 && activityHaversine(lastLat, lastLng, p.Lat, p.Lng) < 8 {
			continue
		}
		if lastLat != 0 {
			elapsed := p.RecordedAt.Sub(lastRecordedAt).Seconds()
			if elapsed > 0 && activityHaversine(lastLat, lastLng, p.Lat, p.Lng)/elapsed > 10 {
				continue
			}
		}
		ps = append(ps, model.ActivityPoint{ActivityID: id, Lat: p.Lat, Lng: p.Lng, AccuracyM: p.AccuracyM, RecordedAt: p.RecordedAt})
		if lastLat != 0 {
			a.DistanceM += activityHaversine(lastLat, lastLng, p.Lat, p.Lng)
		}
		lastLat, lastLng = p.Lat, p.Lng
		lastRecordedAt = p.RecordedAt
	}
	if len(ps) > 0 {
		if e = s.repo.AddPoints(ps); e != nil {
			return nil, e
		}
		a.DurationS = activeDuration(a, time.Now())
		if e = s.repo.Save(a); e != nil {
			return nil, e
		}
	}
	return s.repo.FindByID(id)
}
func (s *activityService) SetStatus(uid, id uuid.UUID, status model.ActivityStatus, occurredAt time.Time) (*model.Activity, error) {
	a, e := s.owned(uid, id)
	if e != nil {
		return nil, e
	}
	if status == model.ActivityStatusActive && a.Status != model.ActivityStatusPaused {
		return nil, errors.New("activity cannot be resumed")
	}
	if status == model.ActivityStatusPaused && a.Status != model.ActivityStatusActive {
		return nil, errors.New("activity cannot be paused")
	}
	if status == model.ActivityStatusCompleted && a.Status != model.ActivityStatusActive && a.Status != model.ActivityStatusPaused {
		return nil, errors.New("activity cannot be completed")
	}
	a.Status = status
	if status == model.ActivityStatusPaused {
		a.DurationS = activeDuration(a, occurredAt)
		a.PausedAt = &occurredAt
	}
	if status == model.ActivityStatusActive && a.PausedAt != nil {
		a.PausedDurationS += int(occurredAt.Sub(*a.PausedAt).Seconds())
		a.PausedAt = nil
		a.DurationS = activeDuration(a, occurredAt)
	}
	if status == model.ActivityStatusCompleted {
		if a.PausedAt != nil {
			a.PausedDurationS += int(occurredAt.Sub(*a.PausedAt).Seconds())
			a.PausedAt = nil
		}
		a.EndedAt = &occurredAt
		a.DurationS = activeDuration(a, occurredAt)
	}
	if e = s.repo.Save(a); e != nil {
		return nil, e
	}
	return s.repo.FindByID(id)
}
func activeDuration(a *model.Activity, at time.Time) int {
	end := at
	if a.EndedAt != nil {
		end = *a.EndedAt
	}
	if a.PausedAt != nil {
		end = *a.PausedAt
	}
	return int(end.Sub(a.StartedAt).Seconds()) - a.PausedDurationS
}
func (s *activityService) Cancel(uid, id uuid.UUID) error {
	a, e := s.owned(uid, id)
	if e != nil {
		return e
	}
	return s.repo.Delete(a.ID)
}
func activityHaversine(a, b, c, d float64) float64 {
	const r = 6371000.
	p := math.Pi / 180
	x := (c - a) * p
	y := (d - b) * p
	v := math.Sin(x/2)*math.Sin(x/2) + math.Cos(a*p)*math.Cos(c*p)*math.Sin(y/2)*math.Sin(y/2)
	return r * 2 * math.Atan2(math.Sqrt(v), math.Sqrt(1-v))
}
