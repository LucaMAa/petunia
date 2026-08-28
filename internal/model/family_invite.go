package model

import (
	"time"

	"github.com/google/uuid"
)

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusDeclined InviteStatus = "declined"
)

type FamilyInvite struct {
	ID        uint         `gorm:"primaryKey"        json:"id"`
	FamilyID  uuid.UUID    `gorm:"type:text;index"   json:"family_id"`
	InviterID uuid.UUID    `gorm:"type:text"         json:"inviter_id"`
	InviteeID uuid.UUID    `gorm:"type:text;index"   json:"invitee_id"`
	Status    InviteStatus `gorm:"type:text;default:'pending'" json:"status"`
	CreatedAt time.Time    `                         json:"created_at"`
	ExpiresAt time.Time    `                         json:"expires_at"`

	Family  *Family `gorm:"foreignKey:FamilyID"  json:"family,omitempty"`
	Inviter *User   `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
}
