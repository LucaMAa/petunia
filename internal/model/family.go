package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FamilyRole string

const (
	FamilyRoleOwner  FamilyRole = "owner"
	FamilyRoleMember FamilyRole = "member"
)

type Family struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"  json:"id"`
	CreatedAt time.Time      `                             json:"created_at"`
	UpdatedAt time.Time      `                             json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                 json:"-"`

	Name         string         `gorm:"not null"              json:"name"`
	AvatarFileID *uuid.UUID     `gorm:"type:text;index"      json:"avatar_file_id,omitempty"`
	Avatar       *UploadedFile  `gorm:"foreignKey:AvatarFileID" json:"avatar,omitempty"`
	Members      []FamilyMember `gorm:"foreignKey:FamilyID"   json:"members,omitempty"`
	Pets         []Pet          `gorm:"many2many:family_pets;" json:"pets,omitempty"`
}

type FamilyMember struct {
	ID       uint       `gorm:"primaryKey"            json:"id"`
	FamilyID uuid.UUID  `gorm:"type:text;index"       json:"family_id"`
	UserID   uuid.UUID  `gorm:"type:text;index"       json:"user_id"`
	Role     FamilyRole `gorm:"type:text;default:'member'" json:"role"`
	JoinedAt time.Time  `                             json:"joined_at"`

	User   *User   `gorm:"foreignKey:UserID"     json:"user,omitempty"`
	Family *Family `gorm:"foreignKey:FamilyID"   json:"-"`
}

func (f *Family) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

func (f *Family) AvatarURL() string {
	if f.Avatar != nil {
		return f.Avatar.URL
	}
	return ""
}
