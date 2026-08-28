package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, envelope{Success: true, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, envelope{Success: false, Error: msg})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, envelope{Success: false, Error: msg})
}

func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, envelope{Success: false, Error: "access denied"})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, envelope{Success: false, Error: msg})
}

func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, envelope{Success: false, Error: msg})
}

func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, envelope{Success: false, Error: err.Error()})
}
