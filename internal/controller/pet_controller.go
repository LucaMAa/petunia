package controller

import (
	"petunia/internal/dto"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PetController struct {
	petSvc service.PetService
}

func NewPetController(petSvc service.PetService) *PetController {
	return &PetController{petSvc: petSvc}
}

func (ctrl *PetController) CreatePet(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input dto.CreatePetDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	pet, err := ctrl.petSvc.CreatePet(uid, input)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, pet)
}

func (ctrl *PetController) GetMyPets(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	pets, err := ctrl.petSvc.GetMyPets(uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, pets)
}

func (ctrl *PetController) GetPet(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}

	pet, err := ctrl.petSvc.GetPet(petID, uid)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, pet)
}

func (ctrl *PetController) UpdatePet(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}

	var input dto.UpdatePetDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	pet, err := ctrl.petSvc.UpdatePet(petID, uid, input)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, pet)
}

func (ctrl *PetController) DeletePet(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}

	if err := ctrl.petSvc.DeletePet(petID, uid); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.NotFound(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (ctrl *PetController) AddOwner(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}

	var input dto.AddOwnerDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.petSvc.AddOwner(petID, uid, input); err != nil {
		switch err.Error() {
		case "access denied":
			response.Forbidden(c)
		case "user is already an owner":
			response.Conflict(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, gin.H{"message": "Owner added"})
}

func (ctrl *PetController) RemoveOwner(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid pet id")
		return
	}

	targetID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := ctrl.petSvc.RemoveOwner(petID, uid, targetID); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.NoContent(c)
}
