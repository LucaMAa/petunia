package controller

import (
	"petunia/internal/dto"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileController struct {
	profileSvc service.ProfileService
}

func NewProfileController(profileSvc service.ProfileService) *ProfileController {
	return &ProfileController{profileSvc: profileSvc}
}

func mustUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, _ := c.Get("user_id")
	uid, ok := raw.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "not authenticated")
	}
	return uid, ok
}

func (ctrl *ProfileController) GetProfile(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	user, err := ctrl.profileSvc.GetProfile(uid)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, user)
}

func (ctrl *ProfileController) UpdateProfile(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input dto.UpdateProfileDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := ctrl.profileSvc.UpdateProfile(uid, input)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, user)
}

func (ctrl *ProfileController) RequestEmailChange(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input dto.RequestEmailChangeDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.profileSvc.RequestEmailChange(uid, input); err != nil {
		if err.Error() == "email already in use" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "We have sent you an email confirmation"})
}

func (ctrl *ProfileController) ConfirmEmailChange(c *gin.Context) {
	var body struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.profileSvc.ConfirmEmailChange(body.Token); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Email updated with success"})
}

func (ctrl *ProfileController) ChangePassword(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input dto.ChangePasswordDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.profileSvc.ChangePassword(uid, input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Password updated with success"})
}

func (ctrl *ProfileController) DisableAccount(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var body struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.profileSvc.DisableAccount(uid, body.Password); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Account disabled"})
}
