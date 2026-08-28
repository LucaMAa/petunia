package service

import (
	"errors"
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/repository"

	"github.com/google/uuid"
)

type PetService interface {
	CreatePet(ownerID uuid.UUID, input dto.CreatePetDto) (*model.Pet, error)
	GetPet(petID uuid.UUID, requesterID uuid.UUID) (*model.Pet, error)
	GetMyPets(ownerID uuid.UUID) ([]model.Pet, error)
	UpdatePet(petID uuid.UUID, requesterID uuid.UUID, input dto.UpdatePetDto) (*model.Pet, error)
	DeletePet(petID uuid.UUID, requesterID uuid.UUID) error
	AddOwner(petID uuid.UUID, requesterID uuid.UUID, input dto.AddOwnerDto) error
	RemoveOwner(petID uuid.UUID, requesterID uuid.UUID, targetUserID uuid.UUID) error
}

type petService struct {
	petRepo  repository.PetRepository
	userRepo repository.UserRepository
}

func NewPetService(petRepo repository.PetRepository, userRepo repository.UserRepository) PetService {
	return &petService{petRepo: petRepo, userRepo: userRepo}
}

func (s *petService) CreatePet(ownerID uuid.UUID, input dto.CreatePetDto) (*model.Pet, error) {
	owner, err := s.userRepo.FindByID(ownerID)
	if err != nil || owner == nil {
		return nil, errors.New("user not found")
	}

	pet := &model.Pet{
		Name:      input.Name,
		Species:   input.Species,
		Breed:     input.Breed,
		BirthDate: input.BirthDate,
		Gender:    input.Gender,
		Owners:    []model.User{*owner},
	}

	if err := s.petRepo.Create(pet); err != nil {
		return nil, err
	}
	return pet, nil
}

func (s *petService) GetPet(petID uuid.UUID, requesterID uuid.UUID) (*model.Pet, error) {
	pet, err := s.petRepo.FindByID(petID)
	if err != nil || pet == nil {
		return nil, errors.New("pet not found")
	}

	visible, err := s.petRepo.FindVisibleToUser(requesterID)
	if err != nil {
		return nil, err
	}
	found := false
	for _, p := range visible {
		if p.ID == petID {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("access denied")
	}

	return pet, nil
}

func (s *petService) GetMyPets(ownerID uuid.UUID) ([]model.Pet, error) {
	return s.petRepo.FindVisibleToUser(ownerID)
}

func (s *petService) UpdatePet(petID uuid.UUID, requesterID uuid.UUID, input dto.UpdatePetDto) (*model.Pet, error) {
	pet, err := s.petRepo.FindByID(petID)
	if err != nil || pet == nil {
		return nil, errors.New("pet not found")
	}

	isOwner, err := s.petRepo.IsOwner(petID, requesterID)
	if err != nil || !isOwner {
		return nil, errors.New("access denied")
	}

	pet.Name = input.Name
	pet.Species = input.Species
	pet.Breed = input.Breed
	pet.BirthDate = input.BirthDate
	pet.Gender = input.Gender

	if err := s.petRepo.Save(pet); err != nil {
		return nil, err
	}
	return pet, nil
}

func (s *petService) DeletePet(petID uuid.UUID, requesterID uuid.UUID) error {
	pet, err := s.petRepo.FindByID(petID)
	if err != nil || pet == nil {
		return errors.New("pet not found")
	}

	isOwner, err := s.petRepo.IsOwner(petID, requesterID)
	if err != nil || !isOwner {
		return errors.New("access denied")
	}

	return s.petRepo.Delete(petID)
}

func (s *petService) AddOwner(petID uuid.UUID, requesterID uuid.UUID, input dto.AddOwnerDto) error {
	pet, err := s.petRepo.FindByID(petID)
	if err != nil || pet == nil {
		return errors.New("pet not found")
	}

	isOwner, err := s.petRepo.IsOwner(petID, requesterID)
	if err != nil || !isOwner {
		return errors.New("access denied")
	}

	targetID, _ := uuid.Parse(input.UserID)
	newOwner, err := s.userRepo.FindByID(targetID)
	if err != nil || newOwner == nil {
		return errors.New("user not found")
	}

	alreadyOwner, _ := s.petRepo.IsOwner(petID, targetID)
	if alreadyOwner {
		return errors.New("user is already an owner")
	}

	return s.petRepo.AddOwner(pet, newOwner)
}

func (s *petService) RemoveOwner(petID uuid.UUID, requesterID uuid.UUID, targetUserID uuid.UUID) error {
	pet, err := s.petRepo.FindByID(petID)
	if err != nil || pet == nil {
		return errors.New("pet not found")
	}

	isOwner, err := s.petRepo.IsOwner(petID, requesterID)
	if err != nil || !isOwner {
		return errors.New("access denied")
	}

	if requesterID == targetUserID && len(pet.Owners) == 1 {
		return errors.New("cannot remove the last owner")
	}

	targetUser, err := s.userRepo.FindByID(targetUserID)
	if err != nil || targetUser == nil {
		return errors.New("user not found")
	}

	return s.petRepo.RemoveOwner(pet, targetUser)
}
