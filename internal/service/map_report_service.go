package service

import (
	"errors"
	"math"
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/repository"
	"petunia/internal/ws"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MapReportService interface {
	CreateReport(userID uuid.UUID, input dto.CreateReportDto) (*dto.ReportResponse, error)
	GetReport(id uuid.UUID) (*dto.ReportResponse, error)
	GetNearby(params dto.NearbyReportsParams) ([]dto.ReportResponse, error)
	GetMyReports(userID uuid.UUID) ([]dto.ReportResponse, error)
	DeleteReport(id uuid.UUID, userID uuid.UUID) error
	ReportAbuse(reportID uuid.UUID, userID uuid.UUID, input dto.ReportAbuseDto) error
}

type mapReportService struct {
	repo     repository.MapReportRepository
	userRepo repository.UserRepository
	geoSvc   GeoService
	pushSvc  PushService
}

func NewMapReportService(repo repository.MapReportRepository, userRepo repository.UserRepository, geoSvc GeoService, pushSvc PushService) MapReportService {
	return &mapReportService{repo: repo, userRepo: userRepo, geoSvc: geoSvc, pushSvc: pushSvc}
}

func (s *mapReportService) CreateReport(userID uuid.UUID, input dto.CreateReportDto) (*dto.ReportResponse, error) {
	var expiresAt *time.Time
	if input.Type == model.ReportTypePoisonedBait || input.Type == model.ReportTypeDanger {
		t := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = &t
	} else {
		t := time.Now().Add(90 * 24 * time.Hour)
		expiresAt = &t
	}

	report := &model.MapReport{
		UserID:      userID,
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		Lat:         input.Lat,
		Lng:         input.Lng,
		ImageURLs:   strings.Join(input.ImageURLs, ","),
		Status:      model.ReportStatusPending,
		ExpiresAt:   expiresAt,
	}

	if err := s.repo.Create(report); err != nil {
		return nil, err
	}

	go s.broadcastToNearbyUsers(report)

	return toResponse(report, 0), nil
}

func (s *mapReportService) GetReport(id uuid.UUID) (*dto.ReportResponse, error) {
	r, err := s.repo.FindByID(id)
	if err != nil || r == nil {
		return nil, errors.New("report not found")
	}
	return toResponse(r, 0), nil
}

func (s *mapReportService) GetNearby(params dto.NearbyReportsParams) ([]dto.ReportResponse, error) {
	radius := params.Radius
	if radius <= 0 || radius > 50000 {
		radius = 2000
	}

	reports, err := s.repo.FindNearby(params.Lat, params.Lng, radius)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ReportResponse, len(reports))
	for i, r := range reports {
		dist := haversine(params.Lat, params.Lng, r.Lat, r.Lng)
		rCopy := r
		result[i] = *toResponse(&rCopy, dist)
	}
	return result, nil
}

func (s *mapReportService) GetMyReports(userID uuid.UUID) ([]dto.ReportResponse, error) {
	reports, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ReportResponse, len(reports))
	for i, r := range reports {
		rCopy := r
		result[i] = *toResponse(&rCopy, 0)
	}
	return result, nil
}

func (s *mapReportService) DeleteReport(id uuid.UUID, userID uuid.UUID) error {
	r, err := s.repo.FindByID(id)
	if err != nil || r == nil {
		return errors.New("report not found")
	}
	if r.UserID != userID {
		return errors.New("access denied")
	}
	return s.repo.Delete(id)
}

func (s *mapReportService) ReportAbuse(reportID uuid.UUID, userID uuid.UUID, input dto.ReportAbuseDto) error {
	r, err := s.repo.FindByID(reportID)
	if err != nil || r == nil {
		return errors.New("report not found")
	}

	already, _ := s.repo.HasAbused(reportID, userID)
	if already {
		return errors.New("already reported")
	}

	abuse := &model.ReportAbuse{
		ReportID: reportID,
		UserID:   userID,
		Reason:   input.Reason,
	}
	if err := s.repo.CreateAbuse(abuse); err != nil {
		return err
	}

	count, _ := s.repo.CountAbuses(reportID)
	r.AbuseCount = int(count)
	if count >= 5 {
		r.Status = model.ReportStatusRejected
	}
	return s.repo.Save(r)
}

func (s *mapReportService) broadcastToNearbyUsers(report *model.MapReport) {
	if report.Type != model.ReportTypePoisonedBait && report.Type != model.ReportTypeDanger {
		return
	}

	ws.GlobalHub.BroadcastNearby(report.Lat, report.Lng, 1000, ws.Message{
		Event: ws.EventNearbyReport,
		Payload: map[string]interface{}{
			"report_id":   report.ID.String(),
			"type":        report.Type,
			"title":       report.Title,
			"description": report.Description,
			"lat":         report.Lat,
			"lng":         report.Lng,
		},
	})
	nearbyUserIDs, _ := s.geoSvc.NearbyUserIDs(report.Lat, report.Lng, 500)
	s.pushSvc.SendToUsers(nearbyUserIDs,
		"☠️ Attenzione: boccone avvelenato",
		report.Title,
		map[string]string{"type": "nearby_report", "report_id": report.ID.String()},
	)
}

func toResponse(r *model.MapReport, distanceM float64) *dto.ReportResponse {
	urls := []string{}
	if r.ImageURLs != "" {
		urls = strings.Split(r.ImageURLs, ",")
	}
	return &dto.ReportResponse{
		MapReport: r,
		ImageURLs: urls,
		Distance:  math.Round(distanceM),
	}
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
