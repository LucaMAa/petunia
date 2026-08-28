package repository

import (
	"petunia/internal/config"
	"petunia/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FamilyInviteRepository interface {
	Create(inv *model.FamilyInvite) error
	FindByID(id uint) (*model.FamilyInvite, error)
	FindPendingByInvitee(inviteeID uuid.UUID) ([]model.FamilyInvite, error)
	FindPendingByFamilyAndInvitee(familyID, inviteeID uuid.UUID) (*model.FamilyInvite, error)
	UpdateStatus(id uint, status model.InviteStatus) error
	DeleteExpired() error
	FindPendingByInviter(inviterID uuid.UUID) ([]model.FamilyInvite, error)
}

type familyInviteRepository struct{ db *gorm.DB }

func NewFamilyInviteRepository() FamilyInviteRepository {
	return &familyInviteRepository{db: config.DB}
}

func (r *familyInviteRepository) Create(inv *model.FamilyInvite) error {
	return r.db.Create(inv).Error
}

func (r *familyInviteRepository) FindByID(id uint) (*model.FamilyInvite, error) {
	var inv model.FamilyInvite
	err := r.db.
		Preload("Family").
		Preload("Inviter").
		First(&inv, id).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *familyInviteRepository) FindPendingByInvitee(inviteeID uuid.UUID) ([]model.FamilyInvite, error) {
	var invites []model.FamilyInvite
	err := r.db.
		Preload("Family").
		Preload("Inviter").
		Where("invitee_id = ? AND status = ? AND expires_at > ?",
			inviteeID, model.InviteStatusPending, time.Now()).
		Find(&invites).Error
	return invites, err
}

func (r *familyInviteRepository) FindPendingByFamilyAndInvitee(familyID, inviteeID uuid.UUID) (*model.FamilyInvite, error) {
	var inv model.FamilyInvite
	err := r.db.Where(
		"family_id = ? AND invitee_id = ? AND status = ? AND expires_at > ?",
		familyID, inviteeID, model.InviteStatusPending, time.Now(),
	).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *familyInviteRepository) UpdateStatus(id uint, status model.InviteStatus) error {
	return r.db.Model(&model.FamilyInvite{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *familyInviteRepository) DeleteExpired() error {
	return r.db.
		Where("expires_at < ? OR status != ?", time.Now(), model.InviteStatusPending).
		Delete(&model.FamilyInvite{}).Error
}

func (r *familyInviteRepository) FindPendingByInviter(inviterID uuid.UUID) ([]model.FamilyInvite, error) {
	var invites []model.FamilyInvite
	err := r.db.
		Preload("Family").
		Preload("Inviter").
		Where("inviter_id = ? AND status = ? AND expires_at > ?",
			inviterID, model.InviteStatusPending, time.Now()).
		Find(&invites).Error
	return invites, err
}
