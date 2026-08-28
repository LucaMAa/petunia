package dto

import "time"

type CreatePetDto struct {
	Name      string     `json:"name"       binding:"required"`
	Species   string     `json:"species"    binding:"required"`
	Breed     string     `json:"breed"`
	BirthDate *time.Time `json:"birth_date"`
	Gender    string     `json:"gender"`
}

type UpdatePetDto struct {
	Name      string     `json:"name"       binding:"required"`
	Species   string     `json:"species"    binding:"required"`
	Breed     string     `json:"breed"`
	BirthDate *time.Time `json:"birth_date" time_format:"2006-01-02"`
	Gender    string     `json:"gender"`
}

type AddOwnerDto struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

type PetResponseDto struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Species   string     `json:"species"`
	Breed     string     `json:"breed"`
	BirthDate *time.Time `json:"birth_date"`
	Gender    string     `json:"gender"`
	AvatarURL string     `json:"avatar_url"`
}
