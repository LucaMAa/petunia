package service

import (
	"errors"
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/repository"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ProfileService interface {
	GetProfile(userID uuid.UUID) (*model.User, error)
	UpdateProfile(userID uuid.UUID, input dto.UpdateProfileDto) (*model.User, error)
	RequestEmailChange(userID uuid.UUID, input dto.RequestEmailChangeDto) error
	ConfirmEmailChange(token string) error
	ChangePassword(userID uuid.UUID, input dto.ChangePasswordDto) error
	DisableAccount(userID uuid.UUID, password string) error
}

type profileService struct {
	userRepo        repository.UserRepository
	emailChangeRepo repository.EmailChangeRepository
}

func NewProfileService(
	userRepo repository.UserRepository,
	emailChangeRepo repository.EmailChangeRepository,
) ProfileService {
	return &profileService{
		userRepo:        userRepo,
		emailChangeRepo: emailChangeRepo,
	}
}

func (s *profileService) GetProfile(userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *profileService) UpdateProfile(userID uuid.UUID, input dto.UpdateProfileDto) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName

	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *profileService) RequestEmailChange(userID uuid.UUID, input dto.RequestEmailChangeDto) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	existing, err := s.userRepo.FindByEmail(input.NewEmail)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("email already in use")
	}

	_ = s.emailChangeRepo.DeleteByUserID(userID.String())

	token := uuid.NewString()
	ec := &model.EmailChange{
		UserID:    userID.String(),
		NewEmail:  input.NewEmail,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := s.emailChangeRepo.Create(ec); err != nil {
		return err
	}

	// go func() {
	// 	if err := mailer.SendEmailChangeConfirmation(input.NewEmail, user.FirstName, input.NewEmail, token); err != nil {
	// 		log.Printf("⚠️  Email change confirmation not sent to %s: %v", input.NewEmail, err)
	// 	}
	// }()

	return nil
}

func (s *profileService) ConfirmEmailChange(token string) error {
	ec, err := s.emailChangeRepo.FindByToken(token)
	if err != nil || ec == nil {
		return errors.New("invalid Token")
	}

	uid, err := uuid.Parse(ec.UserID)
	if err != nil {
		return errors.New("invalid Token")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil || user == nil {
		return errors.New("uuser not found")
	}

	existing, err := s.userRepo.FindByEmail(ec.NewEmail)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != uid {
		return errors.New("email alredy exists")
	}

	user.Email = ec.NewEmail
	if err := s.userRepo.Save(user); err != nil {
		return err
	}

	_ = s.emailChangeRepo.DeleteByUserID(ec.UserID)
	return nil
}

func (s *profileService) ChangePassword(userID uuid.UUID, input dto.ChangePasswordDto) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
		return errors.New("invalid password")
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

func (s *profileService) DisableAccount(userID uuid.UUID, password string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if user.Password != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			return errors.New("invalid password")
		}
	}

	user.Status = model.StatusDisabled
	return s.userRepo.Save(user)
}
