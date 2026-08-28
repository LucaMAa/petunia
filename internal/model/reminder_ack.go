package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderAck struct {
	ID            uuid.UUID `gorm:"type:text;primaryKey"      json:"id"`
	ReminderID    uuid.UUID `gorm:"type:text;index;not null"  json:"reminder_id"`
	OccurrenceKey string    `gorm:"index;not null"            json:"occurrence_key"`
	AckedBy       uuid.UUID `gorm:"type:text;not null"        json:"acked_by"`
	AckedAt       time.Time `                                  json:"acked_at"`

	Reminder *Reminder `gorm:"foreignKey:ReminderID"     json:"reminder,omitempty"`
	User     *User     `gorm:"foreignKey:AckedBy"        json:"user,omitempty"`
}

func (a *ReminderAck) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
