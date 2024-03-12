package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParkingLot struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	DesiredSlots int       `json:"desiredSlots"`
	Slots        []Slot    `json:"slots"`
}

type ParkingLotStatus struct {
	ParkingLotID uuid.UUID    `json:"parkingLotId"`
	Name         string       `json:"name"`
	Slots        []SlotStatus `json:"slots"`
}

type Slot struct {
	ID            uuid.UUID `json:"id"`
	SlotNumber    int       `json:"slotNumber"`
	IsAvailable   bool      `json:"isAvailable"`
	IsMaintenance bool      `json:"isMaintenance"`
}

type SlotStatus struct {
	SlotID          uuid.UUID  `json:"slotId"`
	RegistrationNum *string    `json:"registrationNumber"`
	ParkedAt        *time.Time `json:"parkedAt"`
	UnparkedAt      *time.Time `json:"unparkedAt"`
}

type DailyReport struct {
	TotalVehiclesParked *int `json:"totalVehiclesParked"`
	TotalParkingHours   *int `json:"totalParkingHours"`
	TotalFeeCollected   *int `json:"totalFeeCollected"`
}
