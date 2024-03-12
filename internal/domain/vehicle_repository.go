package domain

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/google/uuid"
)

// VehicleRepository defines the interface for interacting with vehicle data(park, unpark) in the postgresql database.
type VehicleRepository interface {
	ParkVehicle(ctx context.Context, plUUID uuid.UUID, regNum string) (*Vehicle, common.AppError)
	UnparkVehicle(ctx context.Context, regNum string) (*Vehicle, common.AppError)
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
func (v *VehicleRepositoryDB) ParkVehicle(ctx context.Context, plUUID uuid.UUID, regNum string) (*Vehicle, common.AppError) {
	plID, apiErr := getParkingLotIDByUUID(ctx, v.db, v.l, plUUID)
	if apiErr != nil {
		return nil, apiErr
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

	if appErr := v.isVehicleAlreadyParked(ctx, tx, regNum); appErr != nil {
		return nil, appErr
	}

	slotID, slotUUID, appErr := v.findNearestAvailableSlot(ctx, tx, plID, regNum)
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
		RegistrationNumber: regNum,
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
// 1. Existence Check for Vehicle with the Same Registration Number (potential optimization: add an index on registration_number column)
// 2. Retrieves the slotID (int) for efficient querying to availability status update  and slotUUID for client response.
// 3. Executes a query with 'FOR UPDATE'  to lock the nearest available slot, ensuring that concurrent transactions cannot claim the same slot.
// 4. 409 Conflict error if the parking lot is full, 500 Internal Server Error if unexpected database errors occur during the process.
func (v *VehicleRepositoryDB) findNearestAvailableSlot(ctx context.Context, tx *sql.Tx, plID int, regNum string) (int, uuid.UUID, common.AppError) {
	var exists bool
	err := tx.QueryRowContext(ctx, `
       SELECT EXISTS(SELECT 1 FROM vehicles WHERE registration_number = $1 AND unparked_at IS NULL)
    `, regNum).Scan(&exists)

	if err != nil {
		v.l.Error("error checking vehicle existence in the slot", "err", err)
		return 0, uuid.Nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	} else if exists {
		return 0, uuid.Nil, common.NewConflictError("vehicle with this registration number is already parked")
	}

	var slotID int
	var slotUUID uuid.UUID

	err = tx.QueryRowContext(ctx, `
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
		v.l.Info("Chosen nearest slot available", "slot id", slotID, "slot uuid", slotUUID)
		return slotID, slotUUID, nil
	}
}

// UnparkVehicle performs the following within a serializable transaction to guarantee atomicity, and prevent race conditions:
// 1. Finds the parked vehicle using the registration number, ensuring it hasn't already been unparked.
// 2. Calculates the parking fee based on the vehicle's parking duration.
// 3. Updates the vehicle record with the unparking timestamp and calculated fee.
// 4. Marks the corresponding slot as available.
// 5. Returns a Conflict error if the vehicle isn't found or has already been unparked.
// 6. Returns an Internal Server Error if any unexpected database errors occur.
func (v *VehicleRepositoryDB) UnparkVehicle(ctx context.Context, regNum string) (*Vehicle, common.AppError) {
	tx, err := v.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		v.l.Error("error creating unpark vehicle transaction", "err", err)
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

	var vehicle Vehicle
	var slotID int
	err = tx.QueryRowContext(ctx, `
        SELECT uuid, slot_id, parked_at, unparked_at
        FROM vehicles 
        WHERE registration_number = $1 AND unparked_at IS NULL 
        FOR UPDATE`, regNum).Scan(
		&vehicle.ID, &slotID, &vehicle.ParkedAt, &vehicle.UnparkedAt)

	if errors.Is(err, sql.ErrNoRows) {
		v.l.Error("vehicle not found or already unparked", "registration_number", err)
		return nil, common.NewConflictError("vehicle not found or already unparked")
	} else if err != nil {
		v.l.Error("error finding vehicle", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	slotUUID, appErr := getSlotUUIDByID(ctx, tx, v.l, slotID)
	if appErr != nil {
		return nil, appErr
	}
	vehicle.SlotID = slotUUID
	vehicle.RegistrationNumber = regNum

	unparkedAt := time.Now()
	parkingDuration := unparkedAt.Sub(vehicle.ParkedAt)
	hours := int(math.Ceil(parkingDuration.Hours())) // Round up to the nearest hour
	vehicle.Fee = hours * 10
	vehicle.UnparkedAt = &unparkedAt

	_, err = tx.ExecContext(ctx, `
        UPDATE vehicles 
        SET unparked_at = $1
        WHERE uuid = $2`, unparkedAt, vehicle.ID)
	if err != nil {
		v.l.Error("error updating vehicle", "err", err)
		return nil, common.NewInternalServerError("error updating vehicle", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE slots SET is_available = true WHERE id = $1", slotID)
	if err != nil {
		v.l.Error("error updating slot status", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if cmtErr := tx.Commit(); cmtErr != nil {
		v.l.Error("error committing transaction", "err", cmtErr)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, cmtErr)
	}

	return &vehicle, nil
}

// getSlotUUIDByID retrieves the internal integer ID of a slot given its UUID.
func getSlotUUIDByID(ctx context.Context, tx *sql.Tx, l *slog.Logger, slotID int) (uuid.UUID, common.AppError) {
	var slotUUID uuid.UUID
	err := tx.QueryRowContext(ctx, `SELECT uuid FROM slots WHERE id = $1`, slotID).Scan(&slotUUID)
	if errors.Is(err, sql.ErrNoRows) {
		l.Error("slot not found", "err", err)
		return uuid.Nil, common.NewNotFoundError(common.ErrUnexpectedDatabase)
	} else if err != nil {
		l.Error("error fetching slot uuid by id", "err", err)
		return uuid.Nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}
	return slotUUID, nil
}

func (v *VehicleRepositoryDB) isVehicleAlreadyParked(ctx context.Context, tx *sql.Tx, regNum string) common.AppError {
	var existingVehicleID int
	err := tx.QueryRowContext(ctx, `
        SELECT id FROM vehicles 
        WHERE registration_number = $1 
        AND unparked_at IS NULL
    `, regNum).Scan(&existingVehicleID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		v.l.Error("error checking existing vehicle parking status", "err", err)
		return common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return nil
}
