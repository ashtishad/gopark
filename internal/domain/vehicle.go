package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID                 uuid.UUID  `json:"id"`
	RegistrationNumber string     `json:"registrationNumber"`
	SlotID             uuid.UUID  `json:"slotId"`
	ParkedAt           time.Time  `json:"parkedAt"` // park time would be always recorded
	UnparkedAt         *time.Time `json:"unparkedAt,omitempty"`
	Fee                int        `json:"fee,omitempty"`
}
