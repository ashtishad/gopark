package domain

import "github.com/google/uuid"

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
