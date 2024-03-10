package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ParkingLot struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	DesiredSlots int       `json:"desiredSlots"`
	Slots        []Slot    `json:"slots"`
}

type Slot struct {
	ID            uuid.UUID `json:"id"`
	SlotNumber    int       `json:"slotNumber"`
	IsAvailable   bool      `json:"isAvailable"`
	IsMaintenance bool      `json:"isMaintenance"`
}

type SlotStatus struct {
	SlotID          uuid.UUID      `json:"slotId"`
	RegistrationNum sql.NullString `json:"registrationNumber,omitempty"`
	ParkedAt        *time.Time     `json:"parkedAt,omitempty"`
}

type ParkingLotStatus struct {
	ParkingLotID uuid.UUID    `json:"parkingLotId"`
	Name         string       `json:"name"`
	Slots        []SlotStatus `json:"slots"`
}
