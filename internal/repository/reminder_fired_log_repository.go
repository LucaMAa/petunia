package repository

import (
	"errors"
	"petunia/internal/config"
	"petunia/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderFiredLogRepository interface {
	Create(log *model.ReminderFiredLog) error
	FindPendingForUser(userID uuid.UUID) ([]model.ReminderFiredLog, error)
	HasFired(reminderID uuid.UUID, occurrenceKey string) (bool, error)
	DeleteOlderThan(age time.Duration) error
}

type reminderFiredLogRepository struct{ db *gorm.DB }

func NewReminderFiredLogRepository() ReminderFiredLogRepository {
	return &reminderFiredLogRepository{db: config.DB}
}

func (r *reminderFiredLogRepository) Create(l *model.ReminderFiredLog) error {
	return r.db.Create(l).Error
}

func (r *reminderFiredLogRepository) FindPendingForUser(userID uuid.UUID) ([]model.ReminderFiredLog, error) {
	var logs []model.ReminderFiredLog

	cutoff := time.Now().Add(-24 * time.Hour)

	err := r.db.
		Preload("Reminder").
		Preload("Reminder.Pet").
		Joins("JOIN reminders ON reminders.id = reminder_fired_logs.reminder_id").
		Joins("JOIN family_members ON family_members.family_id = reminders.family_id").
		Where(`
			family_members.user_id = ?
			AND reminder_fired_logs.fired_at > ?
			AND reminders.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1 FROM reminder_acks
				WHERE reminder_acks.reminder_id = reminder_fired_logs.reminder_id
				  AND reminder_acks.occurrence_key = reminder_fired_logs.occurrence_key
				  AND reminder_acks.acked_by = ?
			)
		`, userID, cutoff, userID).
		Order("reminder_fired_logs.fired_at DESC").
		Find(&logs).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return logs, err
}

func (r *reminderFiredLogRepository) HasFired(reminderID uuid.UUID, occurrenceKey string) (bool, error) {
	var count int64
	err := r.db.Model(&model.ReminderFiredLog{}).
		Where("reminder_id = ? AND occurrence_key = ?", reminderID, occurrenceKey).
		Count(&count).Error
	return count > 0, err
}

func (r *reminderFiredLogRepository) DeleteOlderThan(age time.Duration) error {
	cutoff := time.Now().Add(-age)
	return r.db.
		Where("fired_at < ?", cutoff).
		Delete(&model.ReminderFiredLog{}).Error
}
