package dto

import (
	"petunia/internal/model"
	"time"

	"github.com/google/uuid"
)

type StartActivityDto struct {
	PetID     *uuid.UUID            `json:"pet_id"`
	Type      model.ActivityType    `json:"type" binding:"required"`
	Privacy   model.ActivityPrivacy `json:"privacy"`
	StartedAt time.Time             `json:"started_at" binding:"required"`
}
type ActivityPointDto struct {
	Lat        float64   `json:"lat" binding:"required"`
	Lng        float64   `json:"lng" binding:"required"`
	AccuracyM  *float64  `json:"accuracy_m"`
	RecordedAt time.Time `json:"recorded_at" binding:"required"`
}
type AppendActivityPointsDto struct {
	Points []ActivityPointDto `json:"points" binding:"required,min=1,max=100"`
}
type ActivityTransitionDto struct {
	OccurredAt time.Time `json:"occurred_at" binding:"required"`
}
