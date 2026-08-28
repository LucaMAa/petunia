package repository

import (
	"petunia/internal/config"
	"petunia/internal/model"
	"time"

	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(r *model.PasswordReset) error
	FindByToken(token string) (*model.PasswordReset, error)
	DeleteByUserID(userID string) error
}

type passwordResetRepository struct{ db *gorm.DB }

func NewPasswordResetRepository() PasswordResetRepository {
	return &passwordResetRepository{db: config.DB}
}

func (r *passwordResetRepository) Create(pr *model.PasswordReset) error {
	return r.db.Create(pr).Error
}

func (r *passwordResetRepository) FindByToken(token string) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&pr).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *passwordResetRepository) DeleteByUserID(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.PasswordReset{}).Error
}
