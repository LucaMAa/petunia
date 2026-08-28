package repository

import (
	"errors"
	"petunia/internal/config"
	"petunia/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PetRepository interface {
	Create(pet *model.Pet) error
	FindByID(id uuid.UUID) (*model.Pet, error)
	FindByOwnerID(ownerID uuid.UUID) ([]model.Pet, error)
	FindVisibleToUser(userID uuid.UUID) ([]model.Pet, error)
	Save(pet *model.Pet) error
	Delete(id uuid.UUID) error
	AddOwner(pet *model.Pet, user *model.User) error
	RemoveOwner(pet *model.Pet, user *model.User) error
	IsOwner(petID uuid.UUID, userID uuid.UUID) (bool, error)
}

type petRepository struct{ db *gorm.DB }

func NewPetRepository() PetRepository {
	return &petRepository{db: config.DB}
}

func (r *petRepository) Create(pet *model.Pet) error {
	return r.db.Create(pet).Error
}

func (r *petRepository) FindByID(id uuid.UUID) (*model.Pet, error) {
	var pet model.Pet
	err := r.db.Preload("Owners").Preload("Avatar").First(&pet, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &pet, err
}

func (r *petRepository) FindByOwnerID(ownerID uuid.UUID) ([]model.Pet, error) {
	var pets []model.Pet
	err := r.db.
		Joins("JOIN pet_owners ON pet_owners.pet_id = pets.id").
		Where("pet_owners.user_id = ? AND pets.deleted_at IS NULL", ownerID).
		Preload("Owners").
		Preload("Avatar").
		Find(&pets).Error
	return pets, err
}

func (r *petRepository) FindVisibleToUser(userID uuid.UUID) ([]model.Pet, error) {
	var pets []model.Pet
	err := r.db.
		Preload("Owners").
		Preload("Avatar").
		Where(`
			pets.deleted_at IS NULL
			AND (
				EXISTS (
					SELECT 1 FROM pet_owners
					WHERE pet_owners.pet_id = pets.id
					  AND pet_owners.user_id = ?
				)
				OR EXISTS (
					SELECT 1 FROM family_pets
					JOIN family_members ON family_members.family_id = family_pets.family_id
					WHERE family_pets.pet_id = pets.id
					  AND family_members.user_id = ?
				)
			)
		`, userID, userID).
		Order("pets.created_at DESC").
		Find(&pets).Error
	return pets, err
}

func (r *petRepository) Save(pet *model.Pet) error {
	return r.db.Save(pet).Error
}

func (r *petRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Pet{}, "id = ?", id).Error
}

func (r *petRepository) AddOwner(pet *model.Pet, user *model.User) error {
	return r.db.Model(pet).Association("Owners").Append(user)
}

func (r *petRepository) RemoveOwner(pet *model.Pet, user *model.User) error {
	return r.db.Model(pet).Association("Owners").Delete(user)
}

func (r *petRepository) IsOwner(petID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("pet_owners").
		Where("pet_id = ? AND user_id = ?", petID, userID).
		Count(&count).Error
	return count > 0, err
}
