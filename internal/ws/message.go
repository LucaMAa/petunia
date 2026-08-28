package ws

const (
	EventFamilyInvite         = "family_invite"
	EventFamilyInviteAccepted = "family_invite_accepted"
	EventFamilyInviteDeclined = "family_invite_declined"
	EventNearbyReport         = "nearby_report"
	EventLocationUpdate       = "location_update"
	EventError                = "error"

	EventReminderFired = "reminder_fired" // cron → all family members
	EventReminderAcked = "reminder_acked" // first ack → all family members
)

type Message struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type FamilyInvitePayload struct {
	InviteID    uint   `json:"invite_id"`
	FamilyID    string `json:"family_id"`
	FamilyName  string `json:"family_name"`
	InviterID   string `json:"inviter_id"`
	InviterName string `json:"inviter_name"`
}

type InviteResponsePayload struct {
	InviteID string `json:"invite_id"`
	Accepted bool   `json:"accepted"`
}

type LocationUpdatePayload struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ReminderFiredPayload struct {
	ReminderID    string `json:"reminder_id"`
	OccurrenceKey string `json:"occurrence_key"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Notes         string `json:"notes"`
	PetName       string `json:"pet_name"`
	FamilyID      string `json:"family_id"`
}

type ReminderAckedPayload struct {
	ReminderID    string `json:"reminder_id"`
	OccurrenceKey string `json:"occurrence_key"`
	Title         string `json:"title"`
	Type          string `json:"type"`
	AckedByName   string `json:"acked_by_name"`
	FamilyID      string `json:"family_id"`
}
