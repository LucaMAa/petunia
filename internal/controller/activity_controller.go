package controller

import (
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ActivityController struct{ svc service.ActivityService }

func NewActivityController(s service.ActivityService) *ActivityController {
	return &ActivityController{s}
}
func (c *ActivityController) Start(ctx *gin.Context) {
	uid, ok := mustUserID(ctx)
	if !ok {
		return
	}
	var in dto.StartActivityDto
	if e := ctx.ShouldBindJSON(&in); e != nil {
		response.BadRequest(ctx, e.Error())
		return
	}
	a, e := c.svc.Start(uid, in)
	if e != nil {
		response.BadRequest(ctx, e.Error())
		return
	}
	response.Created(ctx, a)
}
func (c *ActivityController) List(ctx *gin.Context) {
	uid, ok := mustUserID(ctx)
	if !ok {
		return
	}
	as, e := c.svc.List(uid)
	if e != nil {
		response.InternalError(ctx, e)
		return
	}
	response.OK(ctx, as)
}
func (c *ActivityController) Get(ctx *gin.Context) {
	uid, ok := mustUserID(ctx)
	if !ok {
		return
	}
	id, e := uuid.Parse(ctx.Param("id"))
	if e != nil {
		response.BadRequest(ctx, "invalid id")
		return
	}
	a, e := c.svc.Get(uid, id)
	if e != nil {
		response.NotFound(ctx, e.Error())
		return
	}
	response.OK(ctx, a)
}
func (c *ActivityController) Points(ctx *gin.Context) {
	uid, ok := mustUserID(ctx)
	if !ok {
		return
	}
	id, e := uuid.Parse(ctx.Param("id"))
	if e != nil {
		response.BadRequest(ctx, "invalid id")
		return
	}
	var in dto.AppendActivityPointsDto
	if e = ctx.ShouldBindJSON(&in); e != nil {
		response.BadRequest(ctx, e.Error())
		return
	}
	a, e := c.svc.AddPoints(uid, id, in)
	if e != nil {
		response.BadRequest(ctx, e.Error())
		return
	}
	response.OK(ctx, a)
}
func (c *ActivityController) Status(status model.ActivityStatus) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uid, ok := mustUserID(ctx)
		if !ok {
			return
		}
		id, e := uuid.Parse(ctx.Param("id"))
		if e != nil {
			response.BadRequest(ctx, "invalid id")
			return
		}
		var input dto.ActivityTransitionDto
		if e = ctx.ShouldBindJSON(&input); e != nil {
			response.BadRequest(ctx, e.Error())
			return
		}
		a, e := c.svc.SetStatus(uid, id, status, input.OccurredAt)
		if e != nil {
			response.BadRequest(ctx, e.Error())
			return
		}
		response.OK(ctx, a)
	}
}
func (c *ActivityController) Cancel(ctx *gin.Context) {
	uid, ok := mustUserID(ctx)
	if !ok {
		return
	}
	id, e := uuid.Parse(ctx.Param("id"))
	if e != nil {
		response.BadRequest(ctx, "invalid id")
		return
	}
	if e = c.svc.Cancel(uid, id); e != nil {
		response.BadRequest(ctx, e.Error())
		return
	}
	response.NoContent(ctx)
}
