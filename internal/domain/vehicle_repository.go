package domain

import (
	"context"
	"database/sql"
	"errors"
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
// 2. Parking slots are numbered 1,2,3....n, then we still start from 1 and pick the available one, then Mark this slot as unavailable in the database.
// 3. Creates a new vehicle record associated with the slot and the current UTC timestamp.
// 4. Returns a 409 Conflict error if the parking lot is full.
// 5. Returns a 500 Internal Server Error if unexpected database errors occur during the process.
func (v *VehicleRepositoryDB) ParkVehicle(ctx context.Context, plUUID uuid.UUID, registrationNumber string) (*Vehicle, common.AppError) {
	plID, appErr := getParkingLotIDByUUID(ctx, v.db, v.l, plUUID)
	if appErr != nil {
		return nil, appErr
	}

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

	slotID, slotUUID, appErr := v.findNearestAvailableSlot(ctx, tx, plID)
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

	vehicleInsertQuery := `INSERT INTO vehicles (uuid, registration_number, slot_id, parked_at) VALUES ($1, $2, $3, $4)`
	if _, err = tx.ExecContext(ctx, vehicleInsertQuery, newVehicle.ID, newVehicle.RegistrationNumber, slotID, newVehicle.ParkedAt); err != nil {
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
// 1. Retrieves the slotID (int) for efficient querying to availability status update  and slotUUID for client response.
// 2. Executes a query with 'FOR UPDATE'  to lock the nearest available slot, ensuring that concurrent transactions cannot claim the same slot.
// 3. 409 Conflict error if the parking lot is full, 500 Internal Server Error if unexpected database errors occur during the process.
func (v *VehicleRepositoryDB) findNearestAvailableSlot(ctx context.Context, tx *sql.Tx, plID int) (int, uuid.UUID, common.AppError) {
	var slotID int
	var slotUUID uuid.UUID

	err := tx.QueryRowContext(ctx, `
       SELECT id, uuid FROM slots 
       WHERE parking_lot_id = $1 AND is_available = true AND is_maintenance= false
       ORDER BY slot_number
       LIMIT 1 
       FOR UPDATE`, plID).Scan(&slotID, &slotUUID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		v.l.Error("parking lot is full", "parking_lot_id", plID)
		return 0, uuid.Nil, common.NewConflictError("parking lot is full")
	case err != nil:
		v.l.Error("error finding available slot", "err", err)
		return 0, uuid.Nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	default:
		return slotID, slotUUID, nil
	}
}
