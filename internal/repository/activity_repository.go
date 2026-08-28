package repository

import (
	"errors"
	"petunia/internal/config"
	"petunia/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityRepository interface {
	Create(*model.Activity) error
	FindByID(uuid.UUID) (*model.Activity, error)
	ListByUser(uuid.UUID) ([]model.Activity, error)
	Save(*model.Activity) error
	AddPoints([]model.ActivityPoint) error
	Delete(uuid.UUID) error
}
type activityRepository struct{ db *gorm.DB }

func NewActivityRepository() ActivityRepository              { return &activityRepository{db: config.DB} }
func (r *activityRepository) Create(a *model.Activity) error { return r.db.Create(a).Error }
func (r *activityRepository) FindByID(id uuid.UUID) (*model.Activity, error) {
	var a model.Activity
	err := r.db.Preload("Pet").Preload("Points", func(db *gorm.DB) *gorm.DB { return db.Order("recorded_at") }).First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}
func (r *activityRepository) ListByUser(id uuid.UUID) ([]model.Activity, error) {
	var as []model.Activity
	err := r.db.Preload("Pet").Where("user_id = ?", id).Order("started_at DESC").Find(&as).Error
	return as, err
}
func (r *activityRepository) Save(a *model.Activity) error             { return r.db.Save(a).Error }
func (r *activityRepository) AddPoints(ps []model.ActivityPoint) error { return r.db.Create(&ps).Error }
func (r *activityRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Activity{}, "id = ?", id).Error
}
