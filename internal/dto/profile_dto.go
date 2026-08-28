package dto

type UpdateProfileDto struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
}

type RequestEmailChangeDto struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type ConfirmEmailChangeDto struct {
	Token string `json:"token" binding:"required"`
}
