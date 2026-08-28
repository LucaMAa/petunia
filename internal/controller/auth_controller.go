package controller

import (
	"net/http"
	"os"
	"petunia/internal/dto"
	"petunia/internal/service"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	authSvc service.AuthService
}

func NewAuthController(authSvc service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := os.Getenv("APP_ENV") == "production"
	c.SetCookie("access_token", accessToken, 15*60, "/", "", secure, true)
	c.SetCookie("refresh_token", refreshToken, 7*24*60*60, "/", "", secure, true)
	c.SetSameSite(http.SameSiteLaxMode)
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var input dto.RegisterDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.authSvc.Register(input); err != nil {
		if err.Error() == "email already exist" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var input dto.LoginDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := ctrl.authSvc.Login(input)
	if err != nil {
		switch err.Error() {
		case "account_pending":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_pending"})
		case "account_rejected":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_rejected"})
		case "account_disabled":
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "account_disabled"})
		default:
			response.Unauthorized(c, err.Error())
		}
		return
	}

	setAuthCookies(c, res.Token, res.RefreshToken)
	response.OK(c, gin.H{
		"token":         res.Token,
		"refresh_token": res.RefreshToken,
		"user":          res.User,
	})
}

func (ctrl *AuthController) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := ctrl.authSvc.Refresh(body.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	setAuthCookies(c, res.Token, res.RefreshToken)
	response.OK(c, res)
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.authSvc.Logout(body.RefreshToken); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "logged out"})
}

func (ctrl *AuthController) Google(c *gin.Context) {
	var input dto.GoogleAuthDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := ctrl.authSvc.GoogleLogin(input.Token)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	setAuthCookies(c, res.Token, res.RefreshToken)
	response.OK(c, gin.H{"token": res.Token, "refresh_token": res.RefreshToken, "user": res.User})
}

func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "not authenticated")
		return
	}

	var input dto.ChangePasswordDto
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.authSvc.ChangePassword(userID, input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Password updated with success"})
}

func (ctrl *AuthController) RequestPasswordReset(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = ctrl.authSvc.RequestPasswordReset(body.Email)
	response.OK(c, gin.H{"message": "If email exists you will receive next instructions"})
}

func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var body struct {
		Token       string `json:"token"        binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.authSvc.ResetPassword(body.Token, body.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Password updated with success"})
}
