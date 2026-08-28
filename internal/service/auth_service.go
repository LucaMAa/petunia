package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"petunia/internal/config"
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/repository"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(input dto.RegisterDto) error
	Login(input dto.LoginDto) (*dto.AuthResponseDto, error)
	Refresh(refreshToken string) (*dto.AuthResponseDto, error)
	Logout(refreshToken string) error
	GoogleLogin(token string) (*dto.AuthResponseDto, error)
	ChangePassword(userID uuid.UUID, input dto.ChangePasswordDto) error
	RequestPasswordReset(email string) error
	ResetPassword(token, newPassword string) error
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type authService struct {
	userRepo    repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
	resetRepo   repository.PasswordResetRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	resetRepo repository.PasswordResetRepository,
) AuthService {
	return &authService{userRepo: userRepo, refreshRepo: refreshRepo, resetRepo: resetRepo}
}

func (s *authService) Register(input dto.RegisterDto) error {
	exists, err := s.userRepo.ExistsByEmail(input.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hash),
		Status:    model.StatusEnabled,
	}

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	return nil
}

func (s *authService) Login(input dto.LoginDto) (*dto.AuthResponseDto, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("not valid credentials")
	}

	if user.Status == model.StatusDisabled {
		return nil, errors.New("account_disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, errors.New("not valid credentials")
	}

	accessToken, err := config.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshTokenValue, err := config.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	familyID := uuid.New()
	if err := s.refreshRepo.Create(&model.RefreshToken{
		UserID:    user.ID,
		FamilyID:  familyID,
		TokenHash: hashRefreshToken(refreshTokenValue),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		log.Printf("failed to create refresh token: %v", err)
	}

	return &dto.AuthResponseDto{
		Token:        accessToken,
		RefreshToken: refreshTokenValue,
		User:         user,
	}, nil
}

func (s *authService) Refresh(refreshToken string) (*dto.AuthResponseDto, error) {
	rt, err := s.refreshRepo.FindByToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if rt.RevokedAt != nil {
		if err := s.refreshRepo.RevokeFamily(rt.FamilyID); err != nil {
			log.Printf("failed to revoke family after reuse detection: %v", err)
		}
		return nil, errors.New("refresh token revoked or already used")
	}

	if time.Now().After(rt.ExpiresAt) {
		if err := s.refreshRepo.RevokeFamily(rt.FamilyID); err != nil {
			log.Printf("failed to revoke expired family: %v", err)
		}
		return nil, errors.New("refresh token expired")
	}

	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if err := s.refreshRepo.RevokeToken(refreshToken); err != nil {
		log.Printf("failed to revoke refresh token: %v", err)
	}

	newRefresh, err := config.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	newFamilyID := rt.FamilyID
	newToken := &model.RefreshToken{
		UserID:    user.ID,
		FamilyID:  newFamilyID,
		TokenHash: hashRefreshToken(newRefresh),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshRepo.Create(newToken); err != nil {
		log.Printf("failed to create refresh token: %v", err)
	}

	accessToken, err := config.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponseDto{
		Token:        accessToken,
		RefreshToken: newRefresh,
		User:         user,
	}, nil
}

func (s *authService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("missing refresh token")
	}
	if err := s.refreshRepo.RevokeToken(refreshToken); err != nil {
		return err
	}
	return nil
}

func (s *authService) GoogleLogin(googleToken string) (*dto.AuthResponseDto, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	payload, err := idtoken.Validate(context.Background(), googleToken, clientID)
	if err != nil {
		return nil, errors.New("invalid google token")
	}

	email := payload.Claims["email"].(string)
	firstName, lastName := "", ""
	if v, ok := payload.Claims["given_name"].(string); ok {
		firstName = v
	}
	if v, ok := payload.Claims["family_name"].(string); ok {
		lastName = v
	}
	if firstName == "" {
		if v, ok := payload.Claims["name"].(string); ok {
			firstName = v
		}
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user = &model.User{
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			Status:    model.StatusEnabled,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
		return nil, errors.New("account_pending")
	}

	if user.Status == model.StatusDisabled {
		return nil, errors.New("account_disabled")
	}

	accessToken, err := config.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshTokenValue, err := config.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	familyID := uuid.New()
	if err := s.refreshRepo.Create(&model.RefreshToken{
		UserID:    user.ID,
		FamilyID:  familyID,
		TokenHash: hashRefreshToken(refreshTokenValue),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		log.Printf("failed to create refresh token: %v", err)
	}

	return &dto.AuthResponseDto{Token: accessToken, RefreshToken: refreshTokenValue, User: user}, nil
}

func (s *authService) ChangePassword(userID uuid.UUID, input dto.ChangePasswordDto) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if len(input.NewPassword) < 8 {
		return errors.New("password should be min 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)
	return s.userRepo.Save(user)
}

func (s *authService) RequestPasswordReset(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}
	if user == nil || user.Password == "" {
		return nil
	}

	if err := s.resetRepo.DeleteByUserID(user.ID.String()); err != nil {
		log.Printf("failed to delete previous password reset: %v", err)
	}

	token := uuid.NewString()
	if err := s.resetRepo.Create(&model.PasswordReset{
		UserID:    user.ID.String(),
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}); err != nil {
		log.Printf("failed to create password reset: %v", err)
	}

	// go func() {
	// 	if err := mailer.SendResetPasswordEmail(user.Email, user.FirstName, token); err != nil {
	// 		log.Printf("Reset password email was sent to %s: %v", user.Email, err)
	// 	}
	// }()

	return nil
}

func (s *authService) ResetPassword(token, newPassword string) error {
	pr, err := s.resetRepo.FindByToken(token)
	if err != nil || pr == nil {
		return errors.New("invalid token")
	}

	uid, err := uuid.Parse(pr.UserID)
	if err != nil {
		return errors.New("invalid token")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if len(newPassword) < 8 {
		return errors.New("password should be min 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)
	if err := s.userRepo.Save(user); err != nil {
		return err
	}

	if err := s.resetRepo.DeleteByUserID(pr.UserID); err != nil {
		log.Printf("failed to delete password reset: %v", err)
	}
	return nil
}
