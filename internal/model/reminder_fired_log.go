package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderFiredLog struct {
	ID            uuid.UUID `gorm:"type:text;primaryKey"       json:"id"`
	ReminderID    uuid.UUID `gorm:"type:text;index;not null"   json:"reminder_id"`
	OccurrenceKey string    `gorm:"index;not null"             json:"occurrence_key"`
	FiredAt       time.Time `gorm:"not null"                   json:"fired_at"`
	Reminder      *Reminder `gorm:"foreignKey:ReminderID"      json:"reminder,omitempty"`
}

func (l *ReminderFiredLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
