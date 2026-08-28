package repository

import (
	"errors"
	"petunia/internal/config"
	"petunia/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderRepository interface {
	Create(r *model.Reminder) error
	FindByID(id uuid.UUID) (*model.Reminder, error)
	FindByFamilyID(familyID uuid.UUID) ([]model.Reminder, error)
	FindAllEnabled() ([]model.Reminder, error)
	Save(r *model.Reminder) error
	Delete(id uuid.UUID) error
	CreateAck(a *model.ReminderAck) error
	FindAck(reminderID uuid.UUID, occurrenceKey string) (*model.ReminderAck, error)
	FindByFamilyAndPet(familyID uuid.UUID, petID uuid.UUID) ([]model.Reminder, error)
	FindAcksByFamilyID(familyID uuid.UUID) ([]model.ReminderAck, error)
}

type reminderRepository struct{ db *gorm.DB }

func NewReminderRepository() ReminderRepository {
	return &reminderRepository{db: config.DB}
}

func (r *reminderRepository) Create(rem *model.Reminder) error {
	return r.db.Create(rem).Error
}

func (r *reminderRepository) FindByID(id uuid.UUID) (*model.Reminder, error) {
	var rem model.Reminder
	err := r.db.Preload("Pet").Preload("Family").First(&rem, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rem, err
}

func (r *reminderRepository) FindByFamilyID(familyID uuid.UUID) ([]model.Reminder, error) {
	var list []model.Reminder
	err := r.db.
		Preload("Pet").
		Where("family_id = ? AND deleted_at IS NULL", familyID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *reminderRepository) FindAllEnabled() ([]model.Reminder, error) {
	var list []model.Reminder
	err := r.db.
		Preload("Family").
		Preload("Pet").
		Where("enabled = true AND deleted_at IS NULL").
		Find(&list).Error
	return list, err
}

func (r *reminderRepository) Save(rem *model.Reminder) error {
	return r.db.Save(rem).Error
}

func (r *reminderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Reminder{}, "id = ?", id).Error
}

func (r *reminderRepository) CreateAck(a *model.ReminderAck) error {
	return r.db.Create(a).Error
}

func (r *reminderRepository) FindAck(reminderID uuid.UUID, occurrenceKey string) (*model.ReminderAck, error) {
	var a model.ReminderAck
	err := r.db.
		Preload("User").
		Where("reminder_id = ? AND occurrence_key = ?", reminderID, occurrenceKey).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}

func (r *reminderRepository) FindByFamilyAndPet(familyID uuid.UUID, petID uuid.UUID) ([]model.Reminder, error) {
	var list []model.Reminder
	err := r.db.
		Preload("Pet").
		Where("family_id = ? AND pet_id = ? AND deleted_at IS NULL", familyID, petID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *reminderRepository) FindAcksByFamilyID(familyID uuid.UUID) ([]model.ReminderAck, error) {
	var acks []model.ReminderAck
	err := r.db.
		Joins("JOIN reminders ON reminders.id = reminder_acks.reminder_id").
		Where("reminders.family_id = ? AND reminders.deleted_at IS NULL", familyID).
		Preload("User").
		Preload("Reminder").
		Preload("Reminder.Pet").
		Order("reminder_acks.acked_at DESC").
		Limit(200).
		Find(&acks).Error
	return acks, err
}
