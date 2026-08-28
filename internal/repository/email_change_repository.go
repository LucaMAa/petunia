package repository

import (
	"petunia/internal/config"
	"petunia/internal/model"
	"time"

	"gorm.io/gorm"
)

type EmailChangeRepository interface {
	Create(e *model.EmailChange) error
	FindByToken(token string) (*model.EmailChange, error)
	DeleteByUserID(userID string) error
}

type emailChangeRepository struct{ db *gorm.DB }

func NewEmailChangeRepository() EmailChangeRepository {
	return &emailChangeRepository{db: config.DB}
}

func (r *emailChangeRepository) Create(e *model.EmailChange) error {
	return r.db.Create(e).Error
}

func (r *emailChangeRepository) FindByToken(token string) (*model.EmailChange, error) {
	var e model.EmailChange
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *emailChangeRepository) DeleteByUserID(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.EmailChange{}).Error
}
