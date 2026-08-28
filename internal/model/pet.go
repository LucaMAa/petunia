package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Pet struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"     json:"id"`
	CreatedAt time.Time      `                                json:"created_at"`
	UpdatedAt time.Time      `                                json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                    json:"-"`

	Name         string        `gorm:"not null"                 json:"name"`
	Species      string        `gorm:"not null"                 json:"species"`
	Breed        string        `                                json:"breed"`
	BirthDate    *time.Time    `                                json:"birth_date"`
	Gender       string        `gorm:"type:text"                json:"gender"`
	AvatarFileID *uuid.UUID    `gorm:"type:text;index"          json:"avatar_file_id,omitempty"`
	Avatar       *UploadedFile `gorm:"foreignKey:AvatarFileID"  json:"avatar,omitempty"`

	Owners []User `gorm:"many2many:pet_owners;"    json:"owners,omitempty"`
}

func (p *Pet) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (p *Pet) AvatarURL() string {
	if p.Avatar != nil {
		return p.Avatar.URL
	}
	return ""
}
