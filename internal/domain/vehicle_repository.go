package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/google/uuid"
)

// VehicleRepository defines the interface for interacting with vehicle data(park, unpark) in the postgresql database.
type VehicleRepository interface {
	ParkVehicle(ctx context.Context, parkingLotID uuid.UUID, registrationNumber string, userID uuid.UUID) (*Vehicle, common.AppError)
}

type VehicleRepositoryDB struct {
	db *sql.DB
	l  *slog.Logger
}

func NewVehicleRepoDB(db *sql.DB, l *slog.Logger) *VehicleRepositoryDB {
	return &VehicleRepositoryDB{
		db: db,
		l:  l,
	}
}

// ParkVehicle performs the following within a serializable transaction to guarantee atomicity:
// 1. Locates the nearest available slot in the specified parking lot (using slot numbers) and locks the slot to prevent concurrent updates.
// 2. Marks the selected slot as unavailable in the database.
// 3. Creates a new vehicle record associated with the slot and the current UTC timestamp.
// 4. Returns a 409 Conflict error if the parking lot is full.
// 5. Returns a 500 Internal Server Error if unexpected database errors occur during the process.
func (v *VehicleRepositoryDB) ParkVehicle(ctx context.Context, parkingLotID uuid.UUID, registrationNumber string) (*Vehicle, common.AppError) {
	tx, err := v.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		v.l.Error("error creating park vehicle transaction", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				v.l.Error("error rolling back the transaction", "err", txErr)
			}
		}
	}()

	slotID, slotUUID, appErr := v.findNearestAvailableSlot(ctx, tx, parkingLotID)
	if err != nil {
		return nil, appErr
	}

	_, err = tx.ExecContext(ctx, "UPDATE slots SET is_available = false WHERE id = $1", slotID)
	if err != nil {
		v.l.Error("error updating slot availability status", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	newVehicle := Vehicle{
		ID:                 uuid.New(),
		RegistrationNumber: registrationNumber,
		SlotID:             slotUUID,
		ParkedAt:           time.Now().UTC(),
	}

	err = tx.QueryRowContext(ctx, "INSERT INTO vehicles (vehicle_id, registration_number, slot_id, parked_at) VALUES ($1, $2, $3, $4) RETURNING vehicle_id",
		newVehicle.ID, newVehicle.RegistrationNumber, slotID, newVehicle.ParkedAt).Scan(&newVehicle.ID)
	if err != nil {
		v.l.Error("error creating vehicle record", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if cmtErr := tx.Commit(); err != nil {
		v.l.Error("error committing transaction", "err", cmtErr)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, cmtErr)
	}

	return &newVehicle, nil
}

// findNearestAvailableSlot performs the following steps to locate the nearest vacant slot while preserving data integrity:
// 1. Retrieves the internal ID (integer) of the parking lot based on the provided UUID for efficient database querying.
// 2. Executes a query with 'FOR UPDATE'  to lock the nearest available slot, ensuring that concurrent transactions cannot claim the same slot.
// 3. Returns a 409 Conflict error if the parking lot is full.
// 4. Returns a 500 Internal Server Error if unexpected database errors occur during the process.
// Note: This method returns both the internal slot ID (int) for database interactions and the slot UUID for the client response.
func (v *VehicleRepositoryDB) findNearestAvailableSlot(ctx context.Context, tx *sql.Tx, parkingLotID uuid.UUID) (int, uuid.UUID, common.AppError) {
	var id int
	var slotID uuid.UUID

	var parkingLotInternalID int
	err := tx.QueryRowContext(ctx, `
       SELECT id FROM parking_lots WHERE parking_lot_id = $1`, parkingLotID).Scan(&parkingLotInternalID)
	if err != nil {
		v.l.Error("unable to get parking lot internal id from uuid", "err", err)
		return 0, uuid.Nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	err = tx.QueryRowContext(ctx, `
       SELECT id, slot_id FROM slots 
       WHERE parking_lot_id = $1 AND is_available = true
       ORDER BY slot_number
       LIMIT 1 
       FOR UPDATE`, parkingLotInternalID).Scan(&id, &slotID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		v.l.Error(fmt.Sprintf("parking lot is full at %s", parkingLotID))
		return 0, uuid.Nil, common.NewConflictError("parking lot is full")
	case err != nil:
		v.l.Error("error finding available slot", "err", err)
		return 0, uuid.Nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	default:
		return id, slotID, nil
	}
}
