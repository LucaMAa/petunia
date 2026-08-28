package dto

type CreateFamilyDto struct {
	Name string `json:"name" binding:"required"`
}

type UpdateFamilyDto struct {
	Name string `json:"name" binding:"required"`
}

type InviteMemberDto struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

type AssignPetDto struct {
	PetID string `json:"pet_id" binding:"required,uuid"`
}

type SearchUsersParams struct {
	Query string `form:"q" binding:"required,min=2"`
}
