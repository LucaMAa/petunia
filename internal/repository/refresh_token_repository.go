package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"petunia/internal/config"
	"petunia/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(t *model.RefreshToken) error
	FindByToken(token string) (*model.RefreshToken, error)
	RevokeToken(token string) error
	RevokeFamily(familyID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository() RefreshTokenRepository {
	return &refreshTokenRepository{db: config.DB}
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (r *refreshTokenRepository) Create(t *model.RefreshToken) error {
	return r.db.Create(t).Error
}

func (r *refreshTokenRepository) FindByToken(token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.Where("token_hash = ?", hashRefreshToken(token)).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) RevokeToken(token string) error {
	now := time.Now()
	return r.db.Model(&model.RefreshToken{}).Where("token_hash = ?", hashRefreshToken(token)).Updates(map[string]interface{}{
		"revoked_at": &now,
	}).Error
}

func (r *refreshTokenRepository) RevokeFamily(familyID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&model.RefreshToken{}).Where("family_id = ?", familyID).Updates(map[string]interface{}{
		"revoked_at": &now,
	}).Error
}
