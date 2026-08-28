package repository

import (
	"errors"
	"math"
	"petunia/internal/config"
	"petunia/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GeoIndexer interface {
	AddReport(reportID uuid.UUID, lat, lng float64)
	RemoveReport(reportID uuid.UUID)
	NearbyIDs(lat, lng, radiusMeters float64) ([]string, error)
}

type MapReportRepository interface {
	Create(r *model.MapReport) error
	FindByID(id uuid.UUID) (*model.MapReport, error)
	FindNearby(lat, lng, radiusMeters float64) ([]model.MapReport, error)
	FindByUserID(userID uuid.UUID) ([]model.MapReport, error)
	Save(r *model.MapReport) error
	Delete(id uuid.UUID) error
	CreateAbuse(a *model.ReportAbuse) error
	HasAbused(reportID uuid.UUID, userID uuid.UUID) (bool, error)
	CountAbuses(reportID uuid.UUID) (int64, error)
}

type mapReportRepository struct {
	db  *gorm.DB
	geo GeoIndexer
}

func NewMapReportRepository(geo GeoIndexer) MapReportRepository {
	return &mapReportRepository{db: config.DB, geo: geo}
}

func (r *mapReportRepository) Create(report *model.MapReport) error {
	if err := r.db.Create(report).Error; err != nil {
		return err
	}
	r.geo.AddReport(report.ID, report.Lat, report.Lng)
	return nil
}

func (r *mapReportRepository) FindByID(id uuid.UUID) (*model.MapReport, error) {
	var report model.MapReport
	err := r.db.Preload("User").First(&report, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &report, err
}

func (r *mapReportRepository) FindNearby(lat, lng, radiusMeters float64) ([]model.MapReport, error) {
	ids, err := r.geo.NearbyIDs(lat, lng, radiusMeters)
	if err == nil && len(ids) > 0 {
		return r.findByIDs(ids)
	}
	return r.findNearbyFallback(lat, lng, radiusMeters)
}

func (r *mapReportRepository) findByIDs(ids []string) ([]model.MapReport, error) {
	var reports []model.MapReport
	err := r.db.Preload("User").
		Where("id IN ? AND deleted_at IS NULL AND status != ? AND abuse_count < 10",
			ids, model.ReportStatusRejected).
		Find(&reports).Error
	return reports, err
}

func (r *mapReportRepository) findNearbyFallback(lat, lng, radiusMeters float64) ([]model.MapReport, error) {
	const earthRadius = 6371000.0
	latDelta := radiusMeters / earthRadius * (180 / math.Pi)
	lngDelta := radiusMeters / (earthRadius * math.Cos(lat*math.Pi/180)) * (180 / math.Pi)

	var reports []model.MapReport
	err := r.db.Preload("User").
		Where("deleted_at IS NULL AND status != ? AND abuse_count < 10", model.ReportStatusRejected).
		Where("lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?",
			lat-latDelta, lat+latDelta, lng-lngDelta, lng+lngDelta).
		Find(&reports).Error
	if err != nil {
		return nil, err
	}

	result := reports[:0]
	for _, rep := range reports {
		if haversine(lat, lng, rep.Lat, rep.Lng) <= radiusMeters {
			result = append(result, rep)
		}
	}
	return result, nil
}

func (r *mapReportRepository) FindByUserID(userID uuid.UUID) ([]model.MapReport, error) {
	var reports []model.MapReport
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").Find(&reports).Error
	return reports, err
}

func (r *mapReportRepository) Save(report *model.MapReport) error {
	return r.db.Save(report).Error
}

func (r *mapReportRepository) Delete(id uuid.UUID) error {
	r.geo.RemoveReport(id)
	return r.db.Delete(&model.MapReport{}, "id = ?", id).Error
}

func (r *mapReportRepository) CreateAbuse(a *model.ReportAbuse) error {
	return r.db.Create(a).Error
}

func (r *mapReportRepository) HasAbused(reportID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.ReportAbuse{}).
		Where("report_id = ? AND user_id = ?", reportID, userID).Count(&count).Error
	return count > 0, err
}

func (r *mapReportRepository) CountAbuses(reportID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.ReportAbuse{}).
		Where("report_id = ?", reportID).Count(&count).Error
	return count, err
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
