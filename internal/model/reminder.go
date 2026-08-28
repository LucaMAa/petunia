package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderType string
type ReminderRepeat string

const (
	ReminderTypeMedicine ReminderType = "medicine"
	ReminderTypeFood     ReminderType = "food"
	ReminderTypeOther    ReminderType = "other"

	ReminderRepeatNone   ReminderRepeat = "none"
	ReminderRepeatDaily  ReminderRepeat = "daily"
	ReminderRepeatWeekly ReminderRepeat = "weekly"
	ReminderRepeatCustom ReminderRepeat = "custom"
)

type Reminder struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"        json:"id"`
	CreatedAt time.Time      `                                   json:"created_at"`
	UpdatedAt time.Time      `                                   json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                       json:"-"`

	FamilyID  uuid.UUID  `gorm:"type:text;index;not null"    json:"family_id"`
	PetID     *uuid.UUID `gorm:"type:text;index"             json:"pet_id,omitempty"`
	CreatedBy uuid.UUID  `gorm:"type:text;index;not null"    json:"created_by"`

	Type        ReminderType   `gorm:"type:text;not null"          json:"type"`
	Title       string         `gorm:"not null"                    json:"title"`
	Notes       string         `                                   json:"notes"`
	Repeat      ReminderRepeat `gorm:"type:text;default:'none'"    json:"repeat"`
	CronExpr    string         `gorm:"type:text"                   json:"cron_expr,omitempty"`
	ScheduledAt *time.Time     `                                   json:"scheduled_at,omitempty"`
	TimeOfDay   string         `gorm:"type:text"                   json:"time_of_day,omitempty"`
	DayOfWeek   *int           `                                   json:"day_of_week,omitempty"`
	Enabled     bool           `gorm:"default:true"                json:"enabled"`

	Family *Family `gorm:"foreignKey:FamilyID"         json:"family,omitempty"`
	Pet    *Pet    `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
}

func (r *Reminder) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
