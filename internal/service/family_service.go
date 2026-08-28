package service

import (
	"errors"
	"fmt"
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/repository"
	"petunia/internal/ws"
	"time"

	"github.com/google/uuid"
)

type FamilyService interface {
	CreateFamily(ownerID uuid.UUID, input dto.CreateFamilyDto) (*model.Family, error)
	GetFamily(familyID, requesterID uuid.UUID) (*model.Family, error)
	GetMyFamilies(userID uuid.UUID) ([]model.Family, error)
	UpdateFamily(familyID, requesterID uuid.UUID, input dto.UpdateFamilyDto) (*model.Family, error)
	DeleteFamily(familyID, requesterID uuid.UUID) error
	RemoveMember(familyID, requesterID, targetID uuid.UUID) error
	LeaveFamily(familyID, userID uuid.UUID) error
	AssignPet(familyID, requesterID, petID uuid.UUID) error
	UnassignPet(familyID, requesterID, petID uuid.UUID) error
	SearchMembers(query string, requesterID uuid.UUID) ([]model.User, error)
	SendInvite(familyID, inviterID uuid.UUID, inviteeID uuid.UUID) error
	RespondToInvite(inviteID uint, inviteeID uuid.UUID, accepted bool) error
	GetPendingInvites(userID uuid.UUID) ([]model.FamilyInvite, error)
	CancelInvite(inviteID uint, inviterID uuid.UUID) error
	GetSentInvites(inviterID uuid.UUID) ([]model.FamilyInvite, error)
}

type familyService struct {
	familyRepo repository.FamilyRepository
	userRepo   repository.UserRepository
	petRepo    repository.PetRepository
	inviteRepo repository.FamilyInviteRepository
	pushSvc    PushService
}

func NewFamilyService(
	familyRepo repository.FamilyRepository,
	userRepo repository.UserRepository,
	petRepo repository.PetRepository,
	inviteRepo repository.FamilyInviteRepository,
	pushSvc PushService,
) FamilyService {
	return &familyService{
		familyRepo: familyRepo,
		userRepo:   userRepo,
		petRepo:    petRepo,
		inviteRepo: inviteRepo,
		pushSvc:    pushSvc,
	}
}

func (s *familyService) CreateFamily(ownerID uuid.UUID, input dto.CreateFamilyDto) (*model.Family, error) {
	family := &model.Family{Name: input.Name}
	if err := s.familyRepo.Create(family); err != nil {
		return nil, err
	}

	if err := s.familyRepo.AddMember(family.ID, ownerID, model.FamilyRoleOwner); err != nil {
		return nil, err
	}

	return s.familyRepo.FindByID(family.ID)
}

func (s *familyService) GetFamily(familyID, requesterID uuid.UUID) (*model.Family, error) {
	isMember, err := s.familyRepo.IsMember(familyID, requesterID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	family, err := s.familyRepo.FindByID(familyID)
	if err != nil || family == nil {
		return nil, errors.New("family not found")
	}
	return family, nil
}

func (s *familyService) GetMyFamilies(userID uuid.UUID) ([]model.Family, error) {
	return s.familyRepo.FindByUserID(userID)
}

func (s *familyService) UpdateFamily(familyID, requesterID uuid.UUID, input dto.UpdateFamilyDto) (*model.Family, error) {
	isOwner, err := s.familyRepo.IsOwner(familyID, requesterID)
	if err != nil || !isOwner {
		return nil, errors.New("access denied")
	}

	family, err := s.familyRepo.FindByID(familyID)
	if err != nil || family == nil {
		return nil, errors.New("family not found")
	}

	family.Name = input.Name
	if err := s.familyRepo.Save(family); err != nil {
		return nil, err
	}
	return family, nil
}

func (s *familyService) DeleteFamily(familyID, requesterID uuid.UUID) error {
	isOwner, err := s.familyRepo.IsOwner(familyID, requesterID)
	if err != nil || !isOwner {
		return errors.New("access denied")
	}
	return s.familyRepo.Delete(familyID)
}

func (s *familyService) RemoveMember(familyID, requesterID, targetID uuid.UUID) error {
	isOwner, err := s.familyRepo.IsOwner(familyID, requesterID)
	if err != nil || !isOwner {
		return errors.New("access denied")
	}

	if requesterID == targetID {
		return errors.New("owner cannot remove themselves, delete the family instead")
	}

	isMember, _ := s.familyRepo.IsMember(familyID, targetID)
	if !isMember {
		return errors.New("user is not a member")
	}

	return s.familyRepo.RemoveMember(familyID, targetID)
}

func (s *familyService) LeaveFamily(familyID, userID uuid.UUID) error {
	isOwner, _ := s.familyRepo.IsOwner(familyID, userID)
	if isOwner {
		return errors.New("owner cannot leave, delete the family instead")
	}

	isMember, _ := s.familyRepo.IsMember(familyID, userID)
	if !isMember {
		return errors.New("you are not a member of this family")
	}

	return s.familyRepo.RemoveMember(familyID, userID)
}

func (s *familyService) AssignPet(familyID, requesterID, petID uuid.UUID) error {
	isMember, err := s.familyRepo.IsMember(familyID, requesterID)
	if err != nil || !isMember {
		return errors.New("access denied")
	}

	isOwner, err := s.petRepo.IsOwner(petID, requesterID)
	if err != nil || !isOwner {
		return errors.New("you don't own this pet")
	}

	already, _ := s.familyRepo.HasPet(familyID, petID)
	if already {
		return errors.New("pet already assigned to this family")
	}

	return s.familyRepo.AssignPet(familyID, petID)
}

func (s *familyService) UnassignPet(familyID, requesterID, petID uuid.UUID) error {
	isMember, err := s.familyRepo.IsMember(familyID, requesterID)
	if err != nil || !isMember {
		return errors.New("access denied")
	}

	isOwner, err := s.petRepo.IsOwner(petID, requesterID)
	if err != nil || !isOwner {
		return errors.New("you don't own this pet")
	}

	return s.familyRepo.UnassignPet(familyID, petID)
}

func (s *familyService) SearchMembers(query string, requesterID uuid.UUID) ([]model.User, error) {
	if len(query) < 2 {
		return nil, errors.New("query too short")
	}
	return s.userRepo.SearchByName(query, requesterID)
}

func (s *familyService) SendInvite(familyID, inviterID, inviteeID uuid.UUID) error {
	isMember, err := s.familyRepo.IsMember(familyID, inviterID)
	if err != nil || !isMember {
		return errors.New("access denied")
	}

	invitee, err := s.userRepo.FindByID(inviteeID)
	if err != nil || invitee == nil {
		return errors.New("user not found")
	}

	already, _ := s.familyRepo.IsMember(familyID, inviteeID)
	if already {
		return errors.New("user is already a member")
	}

	existing, _ := s.inviteRepo.FindPendingByFamilyAndInvitee(familyID, inviteeID)
	if existing != nil {
		return errors.New("invite already pending")
	}

	family, err := s.familyRepo.FindByID(familyID)
	if err != nil || family == nil {
		return errors.New("family not found")
	}

	inviter, err := s.userRepo.FindByID(inviterID)
	if err != nil || inviter == nil {
		return errors.New("inviter not found")
	}

	invite := &model.FamilyInvite{
		FamilyID:  familyID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Status:    model.InviteStatusPending,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.inviteRepo.Create(invite); err != nil {
		return err
	}

	ws.GlobalHub.SendToUser(inviteeID, ws.Message{
		Event: ws.EventFamilyInvite,
		Payload: ws.FamilyInvitePayload{
			InviteID:    invite.ID,
			FamilyID:    familyID.String(),
			FamilyName:  family.Name,
			InviterID:   inviterID.String(),
			InviterName: inviter.FirstName + " " + inviter.LastName,
		},
	})

	s.pushSvc.SendToUser(inviteeID.String(),
		"Invito famiglia",
		fmt.Sprintf("%s invited you in %s", inviter.FirstName, family.Name),
		map[string]string{"type": "family_invite", "invite_id": fmt.Sprint(invite.ID)},
	)

	return nil
}

func (s *familyService) RespondToInvite(inviteID uint, inviteeID uuid.UUID, accepted bool) error {
	invite, err := s.inviteRepo.FindByID(inviteID)
	if err != nil || invite == nil {
		return errors.New("invite not found")
	}

	if invite.InviteeID != inviteeID {
		return errors.New("access denied")
	}

	if invite.Status != model.InviteStatusPending {
		return errors.New("invite already answered")
	}

	if time.Now().After(invite.ExpiresAt) {
		return errors.New("invite expired")
	}

	if accepted {
		if err := s.inviteRepo.UpdateStatus(inviteID, model.InviteStatusAccepted); err != nil {
			return err
		}
		if err := s.familyRepo.AddMember(invite.FamilyID, inviteeID, model.FamilyRoleMember); err != nil {
			return err
		}

		ws.GlobalHub.SendToUser(invite.InviterID, ws.Message{
			Event: ws.EventFamilyInviteAccepted,
			Payload: map[string]interface{}{
				"invite_id":   inviteID,
				"family_id":   invite.FamilyID.String(),
				"family_name": invite.Family.Name,
				"invitee_id":  inviteeID.String(),
			},
		})
	} else {
		if err := s.inviteRepo.UpdateStatus(inviteID, model.InviteStatusDeclined); err != nil {
			return err
		}

		ws.GlobalHub.SendToUser(invite.InviterID, ws.Message{
			Event: ws.EventFamilyInviteDeclined,
			Payload: map[string]interface{}{
				"invite_id":  inviteID,
				"family_id":  invite.FamilyID.String(),
				"invitee_id": inviteeID.String(),
			},
		})
	}

	return nil
}

func (s *familyService) GetPendingInvites(userID uuid.UUID) ([]model.FamilyInvite, error) {
	return s.inviteRepo.FindPendingByInvitee(userID)
}

func (s *familyService) CancelInvite(inviteID uint, inviterID uuid.UUID) error {
	invite, err := s.inviteRepo.FindByID(inviteID)
	if err != nil || invite == nil {
		return errors.New("invite not found")
	}
	if invite.InviterID != inviterID {
		return errors.New("access denied")
	}
	return s.inviteRepo.UpdateStatus(inviteID, model.InviteStatusDeclined)
}

func (s *familyService) GetSentInvites(inviterID uuid.UUID) ([]model.FamilyInvite, error) {
	return s.inviteRepo.FindPendingByInviter(inviterID)
}
