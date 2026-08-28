package controller

import (
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
)

type LocationController struct {
	geoSvc service.GeoService
}

func NewLocationController(geoSvc service.GeoService) *LocationController {
	return &LocationController{geoSvc: geoSvc}
}

func (ctrl *LocationController) UpdateLocation(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}

	var input struct {
		Lat float64 `json:"lat" binding:"required"`
		Lng float64 `json:"lng" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctrl.geoSvc.SetUserLocation(uid.String(), input.Lat, input.Lng)
	response.OK(c, gin.H{"message": "location updated"})
}
