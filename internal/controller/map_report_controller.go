package controller

import (
	"petunia/internal/dto"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MapReportController struct {
	svc service.MapReportService
}

func NewMapReportController(svc service.MapReportService) *MapReportController {
	return &MapReportController{svc: svc}
}

func (ctrl *MapReportController) CreateReport(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input dto.CreateReportDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	r, err := ctrl.svc.CreateReport(uid, input)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Created(c, r)
}

func (ctrl *MapReportController) GetNearby(c *gin.Context) {
	var params dto.NearbyReportsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	reports, err := ctrl.svc.GetNearby(params)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, reports)
}

func (ctrl *MapReportController) GetMyReports(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	reports, err := ctrl.svc.GetMyReports(uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, reports)
}

func (ctrl *MapReportController) GetReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	r, err := ctrl.svc.GetReport(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, r)
}

func (ctrl *MapReportController) DeleteReport(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.DeleteReport(id, uid); err != nil {
		if err.Error() == "access denied" {
			response.Forbidden(c)
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.NoContent(c)
}

func (ctrl *MapReportController) ReportAbuse(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var input dto.ReportAbuseDto
	_ = c.ShouldBindJSON(&input)

	if err := ctrl.svc.ReportAbuse(id, uid, input); err != nil {
		switch err.Error() {
		case "already reported":
			response.Conflict(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OK(c, gin.H{"message": "Abuse reported"})
}
