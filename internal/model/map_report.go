package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportType string
type ReportStatus string

const (
	ReportTypePoisonedBait ReportType = "poisoned_bait"
	ReportTypeDogArea      ReportType = "dog_area"
	ReportTypeDanger       ReportType = "danger"
	ReportTypeInteresting  ReportType = "interesting"
	ReportTypeVet          ReportType = "vet"

	ReportStatusPending  ReportStatus = "pending"
	ReportStatusApproved ReportStatus = "approved"
	ReportStatusRejected ReportStatus = "rejected"
)

type MapReport struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"       json:"id"`
	CreatedAt time.Time      `                                  json:"created_at"`
	UpdatedAt time.Time      `                                  json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                      json:"-"`

	UserID      uuid.UUID    `gorm:"type:text;index;not null"   json:"user_id"`
	Type        ReportType   `gorm:"type:text;not null"         json:"type"`
	Title       string       `gorm:"not null"                   json:"title"`
	Description string       `                                  json:"description"`
	Lat         float64      `gorm:"not null"                   json:"lat"`
	Lng         float64      `gorm:"not null"                   json:"lng"`
	ImageURLs   string       `gorm:"type:text"                  json:"-"`
	Status      ReportStatus `gorm:"type:text;default:'pending'" json:"status"`
	AbuseCount  int          `gorm:"default:0"                  json:"abuse_count"`
	ExpiresAt   *time.Time   `                                  json:"expires_at"`

	User *User `gorm:"foreignKey:UserID"          json:"user,omitempty"`
}

func (r *MapReport) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type ReportAbuse struct {
	ID        uint      `gorm:"primaryKey"        json:"id"`
	ReportID  uuid.UUID `gorm:"type:text;index"   json:"report_id"`
	UserID    uuid.UUID `gorm:"type:text;index"   json:"user_id"`
	Reason    string    `gorm:"type:text"         json:"reason"`
	CreatedAt time.Time `                        json:"created_at"`
}
