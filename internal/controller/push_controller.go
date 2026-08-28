package controller

import (
	"petunia/internal/repository"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
)

type PushController struct {
	repo repository.PushTokenRepository
}

func NewPushController(repo repository.PushTokenRepository) *PushController {
	return &PushController{repo: repo}
}

func (ctrl *PushController) Register(c *gin.Context) {
	uid, ok := mustUserID(c)
	if !ok {
		return
	}
	var body struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.repo.Upsert(uid.String(), body.Token, body.Platform); err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "token registered"})
}

func (ctrl *PushController) Unregister(c *gin.Context) {
	var body struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = ctrl.repo.Delete(body.Token)
	response.NoContent(c)
}
