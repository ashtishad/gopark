package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID                 uuid.UUID  `json:"id"`
	RegistrationNumber string     `json:"registrationNumber"`
	SlotID             uuid.UUID  `json:"slotId"`
	ParkedAt           *time.Time `json:"parkedAt"`
	UnparkedAt         *time.Time `json:"unparkedAt"`
	UserID             uuid.UUID  `json:"userId"`
}
