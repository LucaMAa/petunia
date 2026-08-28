package controller

import (
	"petunia/internal/dto"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FamilyController struct {
	familySvc service.FamilyService
}

func NewFamilyController(familySvc service.FamilyService) *FamilyController {
	return &FamilyController{familySvc: familySvc}
}

func (ctrl *FamilyController) CreateFamily(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input dto.CreateFamilyDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	family, err := ctrl.familySvc.CreateFamily(uid, input)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, family)
}

func (ctrl *FamilyController) GetMyFamilies(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	families, err := ctrl.familySvc.GetMyFamilies(uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, families)
}

func (ctrl *FamilyController) GetFamily(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	family, err := ctrl.familySvc.GetFamily(familyID, uid)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, family)
}

func (ctrl *FamilyController) UpdateFamily(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	var input dto.UpdateFamilyDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	family, err := ctrl.familySvc.UpdateFamily(familyID, uid, input)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, family)
}

func (ctrl *FamilyController) DeleteFamily(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	if err := ctrl.familySvc.DeleteFamily(familyID, uid); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.NoContent(c)
}

func (ctrl *FamilyController) RemoveMember(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	targetID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := ctrl.familySvc.RemoveMember(familyID, uid, targetID); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (ctrl *FamilyController) LeaveFamily(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	if err := ctrl.familySvc.LeaveFamily(familyID, uid); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (ctrl *FamilyController) AssignPet(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	var input dto.AssignPetDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	petID, _ := uuid.Parse(input.PetID)

	if err := ctrl.familySvc.AssignPet(familyID, uid, petID); err != nil {
		switch err.Error() {
		case "access denied":
			response.Forbidden(c)
		case "pet already assigned to this family":
			response.Conflict(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, gin.H{"message": "Pet assigned"})
}

func (ctrl *FamilyController) UnassignPet(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	petID, err := uuid.Parse(c.Param("pet_id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}

	if err := ctrl.familySvc.UnassignPet(familyID, uid, petID); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (ctrl *FamilyController) SearchUsers(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	q := c.Query("q")
	if len(q) < 2 {
		response.BadRequest(c, "query must be at least 2 characters")
		return
	}

	users, err := ctrl.familySvc.SearchMembers(q, uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, users)
}

func (ctrl *FamilyController) SendInvite(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	var input dto.InviteMemberDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inviteeID, err := uuid.Parse(input.UserID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := ctrl.familySvc.SendInvite(familyID, uid, inviteeID); err != nil {
		switch err.Error() {
		case "access denied":
			response.Forbidden(c)
		case "user is already a member", "invite already pending":
			response.Conflict(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, gin.H{"message": "Invite sent"})
}

func (ctrl *FamilyController) RespondToInvite(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var body struct {
		InviteID uint `json:"invite_id" binding:"required"`
		Accepted bool `json:"accepted"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.familySvc.RespondToInvite(body.InviteID, uid, body.Accepted); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Response recorded"})
}

func (ctrl *FamilyController) GetPendingInvites(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	invites, err := ctrl.familySvc.GetPendingInvites(uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, invites)
}

func (ctrl *FamilyController) CancelInvite(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var body struct {
		InviteID uint `json:"invite_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.familySvc.CancelInvite(body.InviteID, uid); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (ctrl *FamilyController) GetSentInvites(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}
	invites, err := ctrl.familySvc.GetSentInvites(uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, invites)
}
