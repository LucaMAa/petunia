package controller

import (
	"context"
	"petunia/internal/service"

	"github.com/gin-gonic/gin"
)

type TranslationController struct {
	svc *service.TranslationService
}

func NewTranslationController(s *service.TranslationService) *TranslationController {
	return &TranslationController{svc: s}
}

func (tc *TranslationController) GetTranslations(c *gin.Context) {
	locale := c.Query("locale")
	if locale == "" {
		locale = c.GetHeader("Accept-Language")
		if locale == "" {
			locale = "en-US"
		}
	}

	ctx := context.Background()
	m, err := tc.svc.GetTranslations(ctx, locale)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": m})
}
