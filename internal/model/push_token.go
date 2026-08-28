package model

import "time"

type PushToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    string `gorm:"index;not null"`
	Token     string `gorm:"uniqueIndex;not null"`
	Platform  string `gorm:"type:text"` // "ios" | "android"
	CreatedAt time.Time
	UpdatedAt time.Time
}
