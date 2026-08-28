package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityType string
type ActivityStatus string
type ActivityPrivacy string

const (
	ActivityStatusActive     ActivityStatus  = "active"
	ActivityStatusPaused     ActivityStatus  = "paused"
	ActivityStatusCompleted  ActivityStatus  = "completed"
	ActivityStatusCancelled  ActivityStatus  = "cancelled"
	ActivityPrivacyPrivate   ActivityPrivacy = "private"
	ActivityPrivacyStatsOnly ActivityPrivacy = "stats_only"
	ActivityPrivacyShared    ActivityPrivacy = "shared"
)

type Activity struct {
	ID              uuid.UUID       `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
	UserID          uuid.UUID       `gorm:"type:text;index;not null" json:"user_id"`
	PetID           *uuid.UUID      `gorm:"type:text;index" json:"pet_id,omitempty"`
	Type            ActivityType    `gorm:"type:text;not null" json:"type"`
	Status          ActivityStatus  `gorm:"type:text;index;not null" json:"status"`
	Privacy         ActivityPrivacy `gorm:"type:text;not null;default:'private'" json:"privacy"`
	StartedAt       time.Time       `gorm:"index;not null" json:"started_at"`
	EndedAt         *time.Time      `json:"ended_at,omitempty"`
	PausedAt        *time.Time      `json:"paused_at,omitempty"`
	PausedDurationS int             `gorm:"default:0" json:"paused_duration_s"`
	DurationS       int             `gorm:"default:0" json:"duration_s"`
	DistanceM       float64         `gorm:"default:0" json:"distance_m"`
	Pet             *Pet            `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	Points          []ActivityPoint `gorm:"foreignKey:ActivityID" json:"points,omitempty"`
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type ActivityPoint struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ActivityID uuid.UUID `gorm:"type:text;index;not null" json:"-"`
	Lat        float64   `gorm:"not null" json:"lat"`
	Lng        float64   `gorm:"not null" json:"lng"`
	AccuracyM  *float64  `json:"accuracy_m,omitempty"`
	RecordedAt time.Time `gorm:"index;not null" json:"recorded_at"`
}
