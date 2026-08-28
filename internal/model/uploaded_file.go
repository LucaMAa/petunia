package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileOwnerType string

const (
	FileOwnerPet    FileOwnerType = "pet"
	FileOwnerUser   FileOwnerType = "user"
	FileOwnerFamily FileOwnerType = "family"
)

type FileCategory string

const (
	FileCategoryAvatar   FileCategory = "avatar"
	FileCategoryDocument FileCategory = "document"
)

type UploadedFile struct {
	ID           uuid.UUID      `gorm:"type:text;primaryKey"        json:"id"`
	CreatedAt    time.Time      `                                   json:"created_at"`
	UpdatedAt    time.Time      `                                   json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                       json:"-"`
	UploaderID   uuid.UUID      `gorm:"type:text;index;not null"    json:"uploader_id"`
	OwnerID      uuid.UUID      `gorm:"type:text;index;not null"    json:"owner_id"`
	OwnerType    FileOwnerType  `gorm:"type:text;not null"          json:"owner_type"`
	Category     FileCategory   `gorm:"type:text;not null;default:'document'" json:"category"`
	OriginalName string         `gorm:"type:text;not null"          json:"original_name"`
	MimeType     string         `gorm:"type:text;not null"          json:"mime_type"`
	Size         int64          `gorm:"not null;default:0"          json:"size"`
	StoragePath  string         `gorm:"type:text;not null"          json:"-"`
	URL          string         `gorm:"type:text"                   json:"url"`
}

func (f *UploadedFile) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
