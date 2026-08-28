package dto

import (
	"petunia/internal/model"
	"time"
)

type CreateReminderDto struct {
	FamilyID    string               `json:"family_id"    binding:"required,uuid"`
	PetID       string               `json:"pet_id"`
	Type        model.ReminderType   `json:"type"         binding:"required"`
	Title       string               `json:"title"        binding:"required"`
	Notes       string               `json:"notes"`
	Repeat      model.ReminderRepeat `json:"repeat"`
	CronExpr    string               `json:"cron_expr"`
	ScheduledAt *time.Time           `json:"scheduled_at"`
	TimeOfDay   string               `json:"time_of_day"`
	DayOfWeek   *int                 `json:"day_of_week"`
}

type UpdateReminderDto struct {
	Title       string               `json:"title"        binding:"required"`
	Notes       string               `json:"notes"`
	Repeat      model.ReminderRepeat `json:"repeat"`
	CronExpr    string               `json:"cron_expr"`
	ScheduledAt *time.Time           `json:"scheduled_at"`
	TimeOfDay   string               `json:"time_of_day"`
	DayOfWeek   *int                 `json:"day_of_week"`
	Enabled     bool                 `json:"enabled"`
}

type AckReminderDto struct {
	OccurrenceKey string `json:"occurrence_key" binding:"required"`
}

type GetRemindersParams struct {
	PetID string `form:"pet_id"`
}
