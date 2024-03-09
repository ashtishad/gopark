package domain

import "github.com/google/uuid"

type ParkingLot struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
