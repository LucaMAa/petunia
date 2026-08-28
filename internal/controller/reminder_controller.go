package controller

import (
	"petunia/internal/dto"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReminderController struct {
	svc service.ReminderService
}

func NewReminderController(svc service.ReminderService) *ReminderController {
	return &ReminderController{svc: svc}
}

func (ctrl *ReminderController) Create(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input dto.CreateReminderDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rem, err := ctrl.svc.Create(uid, input)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, rem)
}

func (ctrl *ReminderController) GetByFamily(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var petID *uuid.UUID
	if raw := c.Query("pet_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "invalid pet_id")
			return
		}
		petID = &parsed
	}

	list, err := ctrl.svc.GetByFamily(familyID, uid, petID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, list)
}

func (ctrl *ReminderController) Update(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var input dto.UpdateReminderDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rem, err := ctrl.svc.Update(id, uid, input)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, rem)
}

func (ctrl *ReminderController) Delete(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(id, uid); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.NotFound(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (ctrl *ReminderController) Ack(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var input dto.AckReminderDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	isFirst, err := ctrl.svc.Ack(id, uid, input.OccurrenceKey)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"first_ack": isFirst,
		"message":   "acknowledged",
	})
}

func (ctrl *ReminderController) GetAckHistory(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	familyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid family id")
		return
	}

	acks, err := ctrl.svc.GetAcksByFamily(familyID, uid)
	if err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, acks)
}

func (ctrl *ReminderController) GetPendingAlerts(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	logs, err := ctrl.svc.GetPendingAlerts(uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, logs)
}
