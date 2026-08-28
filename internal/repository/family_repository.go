package repository

import (
	"errors"
	"petunia/internal/config"
	"petunia/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FamilyRepository interface {
	Create(f *model.Family) error
	FindByID(id uuid.UUID) (*model.Family, error)
	FindByUserID(userID uuid.UUID) ([]model.Family, error)
	Save(f *model.Family) error
	Delete(id uuid.UUID) error

	AddMember(familyID, userID uuid.UUID, role model.FamilyRole) error
	RemoveMember(familyID, userID uuid.UUID) error
	FindMember(familyID, userID uuid.UUID) (*model.FamilyMember, error)
	IsOwner(familyID, userID uuid.UUID) (bool, error)
	IsMember(familyID, userID uuid.UUID) (bool, error)

	AssignPet(familyID, petID uuid.UUID) error
	UnassignPet(familyID, petID uuid.UUID) error
	HasPet(familyID, petID uuid.UUID) (bool, error)
}

type familyRepository struct{ db *gorm.DB }

func NewFamilyRepository() FamilyRepository {
	return &familyRepository{db: config.DB}
}

func (r *familyRepository) Create(f *model.Family) error {
	return r.db.Create(f).Error
}

func (r *familyRepository) FindByID(id uuid.UUID) (*model.Family, error) {
	var f model.Family
	err := r.db.
		Preload("Members").
		Preload("Members.User").
		Preload("Pets").
		First(&f, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &f, err
}

func (r *familyRepository) FindByUserID(userID uuid.UUID) ([]model.Family, error) {
	var families []model.Family
	err := r.db.
		Joins("JOIN family_members ON family_members.family_id = families.id").
		Where("family_members.user_id = ? AND families.deleted_at IS NULL", userID).
		Preload("Members").
		Preload("Pets").
		Preload("Members.User").
		Find(&families).Error
	return families, err
}

func (r *familyRepository) Save(f *model.Family) error {
	return r.db.Save(f).Error
}

func (r *familyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Family{}, "id = ?", id).Error
}

func (r *familyRepository) AddMember(familyID, userID uuid.UUID, role model.FamilyRole) error {
	m := model.FamilyMember{
		FamilyID: familyID,
		UserID:   userID,
		Role:     role,
	}
	return r.db.Create(&m).Error
}

func (r *familyRepository) RemoveMember(familyID, userID uuid.UUID) error {
	return r.db.
		Where("family_id = ? AND user_id = ?", familyID, userID).
		Delete(&model.FamilyMember{}).Error
}

func (r *familyRepository) FindMember(familyID, userID uuid.UUID) (*model.FamilyMember, error) {
	var m model.FamilyMember
	err := r.db.Where("family_id = ? AND user_id = ?", familyID, userID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

func (r *familyRepository) IsOwner(familyID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.FamilyMember{}).
		Where("family_id = ? AND user_id = ? AND role = ?", familyID, userID, model.FamilyRoleOwner).
		Count(&count).Error
	return count > 0, err
}

func (r *familyRepository) IsMember(familyID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.FamilyMember{}).
		Where("family_id = ? AND user_id = ?", familyID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *familyRepository) AssignPet(familyID, petID uuid.UUID) error {
	family := &model.Family{}
	family.ID = familyID
	pet := &model.Pet{}
	pet.ID = petID
	return r.db.Model(family).Association("Pets").Append(pet)
}

func (r *familyRepository) UnassignPet(familyID, petID uuid.UUID) error {
	family := &model.Family{}
	family.ID = familyID
	pet := &model.Pet{}
	pet.ID = petID
	return r.db.Model(family).Association("Pets").Delete(pet)
}

func (r *familyRepository) HasPet(familyID, petID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("family_pets").
		Where("family_id = ? AND pet_id = ?", familyID, petID).
		Count(&count).Error
	return count > 0, err
}
