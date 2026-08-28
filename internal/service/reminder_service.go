package service

import (
	"errors"
	"petunia/internal/dto"
	"petunia/internal/model"
	"petunia/internal/repository"
	"petunia/internal/ws"
	"time"

	"github.com/google/uuid"
)

type ReminderService interface {
	Create(userID uuid.UUID, input dto.CreateReminderDto) (*model.Reminder, error)
	GetByFamily(familyID, requesterID uuid.UUID, petID *uuid.UUID) ([]model.Reminder, error)
	Update(id, requesterID uuid.UUID, input dto.UpdateReminderDto) (*model.Reminder, error)
	Delete(id, requesterID uuid.UUID) error
	Ack(reminderID uuid.UUID, userID uuid.UUID, occurrenceKey string) (bool, error)
	FireReminder(reminder *model.Reminder, occurrenceKey string)
	GetAcksByFamily(familyID, requesterID uuid.UUID) ([]model.ReminderAck, error)
	GetPendingAlerts(userID uuid.UUID) ([]model.ReminderFiredLog, error)
}

type reminderService struct {
	reminderRepo repository.ReminderRepository
	familyRepo   repository.FamilyRepository
	userRepo     repository.UserRepository
	firedLogRepo repository.ReminderFiredLogRepository
	pushSvc      PushService
}

func NewReminderService(
	reminderRepo repository.ReminderRepository,
	familyRepo repository.FamilyRepository,
	userRepo repository.UserRepository,
	firedLogRepo repository.ReminderFiredLogRepository,
	pushSvc PushService,
) ReminderService {
	return &reminderService{
		reminderRepo: reminderRepo,
		familyRepo:   familyRepo,
		userRepo:     userRepo,
		firedLogRepo: firedLogRepo,
		pushSvc:      pushSvc,
	}
}

func (s *reminderService) Create(userID uuid.UUID, input dto.CreateReminderDto) (*model.Reminder, error) {
	familyID, err := uuid.Parse(input.FamilyID)
	if err != nil {
		return nil, errors.New("invalid family_id")
	}

	isMember, err := s.familyRepo.IsMember(familyID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	rem := &model.Reminder{
		FamilyID:    familyID,
		CreatedBy:   userID,
		Type:        input.Type,
		Title:       input.Title,
		Notes:       input.Notes,
		Repeat:      input.Repeat,
		CronExpr:    input.CronExpr,
		ScheduledAt: input.ScheduledAt,
		TimeOfDay:   input.TimeOfDay,
		DayOfWeek:   input.DayOfWeek,
		Enabled:     true,
	}

	if input.PetID != "" {
		pid, err := uuid.Parse(input.PetID)
		if err == nil {
			rem.PetID = &pid
		}
	}

	if input.Repeat == "" {
		rem.Repeat = model.ReminderRepeatNone
	}

	if err := s.reminderRepo.Create(rem); err != nil {
		return nil, err
	}
	return rem, nil
}

func (s *reminderService) GetByFamily(familyID, requesterID uuid.UUID, petID *uuid.UUID) ([]model.Reminder, error) {
	isMember, err := s.familyRepo.IsMember(familyID, requesterID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}
	if petID != nil {
		return s.reminderRepo.FindByFamilyAndPet(familyID, *petID)
	}
	return s.reminderRepo.FindByFamilyID(familyID)
}

func (s *reminderService) Update(id, requesterID uuid.UUID, input dto.UpdateReminderDto) (*model.Reminder, error) {
	rem, err := s.reminderRepo.FindByID(id)
	if err != nil || rem == nil {
		return nil, errors.New("reminder not found")
	}

	isMember, err := s.familyRepo.IsMember(rem.FamilyID, requesterID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	rem.Title = input.Title
	rem.Notes = input.Notes
	rem.Repeat = input.Repeat
	rem.CronExpr = input.CronExpr
	rem.ScheduledAt = input.ScheduledAt
	rem.TimeOfDay = input.TimeOfDay
	rem.DayOfWeek = input.DayOfWeek
	rem.Enabled = input.Enabled

	if err := s.reminderRepo.Save(rem); err != nil {
		return nil, err
	}
	return rem, nil
}

func (s *reminderService) Delete(id, requesterID uuid.UUID) error {
	rem, err := s.reminderRepo.FindByID(id)
	if err != nil || rem == nil {
		return errors.New("reminder not found")
	}

	isMember, err := s.familyRepo.IsMember(rem.FamilyID, requesterID)
	if err != nil || !isMember {
		return errors.New("access denied")
	}

	return s.reminderRepo.Delete(id)
}

func (s *reminderService) Ack(reminderID uuid.UUID, userID uuid.UUID, occurrenceKey string) (bool, error) {
	rem, err := s.reminderRepo.FindByID(reminderID)
	if err != nil || rem == nil {
		return false, errors.New("reminder not found")
	}

	isMember, err := s.familyRepo.IsMember(rem.FamilyID, userID)
	if err != nil || !isMember {
		return false, errors.New("access denied")
	}

	existing, err := s.reminderRepo.FindAck(reminderID, occurrenceKey)
	if err != nil {
		return false, err
	}
	if existing != nil {
		return false, nil
	}

	ack := &model.ReminderAck{
		ReminderID:    reminderID,
		OccurrenceKey: occurrenceKey,
		AckedBy:       userID,
		AckedAt:       time.Now(),
	}
	if err := s.reminderRepo.CreateAck(ack); err != nil {
		return false, err
	}

	user, _ := s.userRepo.FindByID(userID)
	userName := "Un membro"
	if user != nil {
		userName = user.FirstName + " " + user.LastName
	}

	s.broadcastAck(rem, occurrenceKey, userName)
	return true, nil
}

func (s *reminderService) FireReminder(reminder *model.Reminder, occurrenceKey string) {
	petName := ""
	if reminder.Pet != nil {
		petName = reminder.Pet.Name
	}

	payload := map[string]interface{}{
		"reminder_id":    reminder.ID.String(),
		"occurrence_key": occurrenceKey,
		"type":           reminder.Type,
		"title":          reminder.Title,
		"notes":          reminder.Notes,
		"pet_name":       petName,
		"family_id":      reminder.FamilyID.String(),
	}

	s.broadcastToFamily(reminder.FamilyID, ws.Message{
		Event:   ws.EventReminderFired,
		Payload: payload,
	})
}

func (s *reminderService) GetPendingAlerts(userID uuid.UUID) ([]model.ReminderFiredLog, error) {
	return s.firedLogRepo.FindPendingForUser(userID)
}

func (s *reminderService) broadcastAck(rem *model.Reminder, occurrenceKey, userName string) {
	payload := map[string]interface{}{
		"reminder_id":    rem.ID.String(),
		"occurrence_key": occurrenceKey,
		"title":          rem.Title,
		"type":           rem.Type,
		"acked_by_name":  userName,
		"family_id":      rem.FamilyID.String(),
	}
	s.broadcastToFamily(rem.FamilyID, ws.Message{
		Event:   ws.EventReminderAcked,
		Payload: payload,
	})
}

func (s *reminderService) broadcastToFamily(familyID uuid.UUID, msg ws.Message) {
	family, err := s.familyRepo.FindByID(familyID)
	if err != nil || family == nil {
		return
	}
	var userIDs []string
	for _, m := range family.Members {
		ws.GlobalHub.SendToUser(m.UserID, msg)
		userIDs = append(userIDs, m.UserID.String())
	}

	if msg.Event == ws.EventReminderFired {
		payload := msg.Payload.(map[string]interface{})
		s.pushSvc.SendToUsers(userIDs, "🔔 "+payload["title"].(string), "Promemoria da fare", map[string]string{"type": "reminder_fired"})
	}
	if msg.Event == ws.EventReminderAcked {
		payload := msg.Payload.(map[string]interface{})
		s.pushSvc.SendToUsers(userIDs, "✓ Fatto", payload["title"].(string)+" segnato come completato", map[string]string{"type": "reminder_acked"})
	}
}

func (s *reminderService) GetAcksByFamily(familyID, requesterID uuid.UUID) ([]model.ReminderAck, error) {
	isMember, err := s.familyRepo.IsMember(familyID, requesterID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}
	return s.reminderRepo.FindAcksByFamilyID(familyID)
}
