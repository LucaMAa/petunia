package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	StatusEnabled  UserStatus = "enabled"
	StatusDisabled UserStatus = "disabled"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"         json:"id"`
	CreatedAt time.Time      `                                    json:"created_at"`
	UpdatedAt time.Time      `                                    json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                        json:"-"`

	FirstName    string        `gorm:"not null"                     json:"first_name"`
	LastName     string        `gorm:"not null"                     json:"last_name"`
	Email        string        `gorm:"uniqueIndex;not null"         json:"email"`
	Password     string        `gorm:"not null"                     json:"-"`
	Status       UserStatus    `gorm:"type:text;default:'approved'" json:"status"`
	AvatarFileID *uuid.UUID    `gorm:"type:text;index"              json:"avatar_file_id,omitempty"`
	Avatar       *UploadedFile `gorm:"foreignKey:AvatarFileID"      json:"avatar,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) AvatarURL() string {
	if u.Avatar != nil {
		return u.Avatar.URL
	}
	return ""
}
