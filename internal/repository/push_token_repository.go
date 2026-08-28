package repository

import (
	"petunia/internal/config"
	"petunia/internal/model"

	"gorm.io/gorm"
)

type PushTokenRepository interface {
	Upsert(userID, token, platform string) error
	Delete(token string) error
	FindByUserID(userID string) ([]model.PushToken, error)
	FindByUserIDs(userIDs []string) ([]model.PushToken, error)
}

type pushTokenRepository struct{ db *gorm.DB }

func NewPushTokenRepository() PushTokenRepository {
	return &pushTokenRepository{db: config.DB}
}

func (r *pushTokenRepository) Upsert(userID, token, platform string) error {
	var existing model.PushToken
	err := r.db.Where("token = ?", token).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&model.PushToken{UserID: userID, Token: token, Platform: platform}).Error
	}
	if err != nil {
		return err
	}
	existing.UserID = userID
	existing.Platform = platform
	return r.db.Save(&existing).Error
}

func (r *pushTokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.PushToken{}).Error
}

func (r *pushTokenRepository) FindByUserID(userID string) ([]model.PushToken, error) {
	var tokens []model.PushToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

func (r *pushTokenRepository) FindByUserIDs(userIDs []string) ([]model.PushToken, error) {
	var tokens []model.PushToken
	if len(userIDs) == 0 {
		return tokens, nil
	}
	err := r.db.Where("user_id IN ?", userIDs).Find(&tokens).Error
	return tokens, err
}
