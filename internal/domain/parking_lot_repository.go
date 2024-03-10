package domain

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/google/uuid"
)

// ParkingLotRepository defines the interface for interacting with parking lot data in the postgresql database.
type ParkingLotRepository interface {
	CreateParkingLot(ctx context.Context, lot *ParkingLot) (*ParkingLot, common.AppError)
}

type ParkingLotRepoDB struct {
	db *sql.DB
	l  *slog.Logger
}

func NewParkingLotRepoDB(db *sql.DB, l *slog.Logger) *ParkingLotRepoDB {
	return &ParkingLotRepoDB{
		db: db,
		l:  l,
	}
}

// CreateParkingLot performs the following within a serializable transaction to ensure consistency:
// 1. Verifies uniqueness of the parking lot name to prevent duplicates (returning a 409 Conflict error if a duplicate exists).
// 2. Inserts the new parking lot record into the database.
// 3. Creates multiple slots associated with the parking lot, using incrementing slot numbers.
// 4. Returns a 500 Internal Server Error if unexpected database errors occur during the process.
func (r *ParkingLotRepoDB) CreateParkingLot(ctx context.Context, lot *ParkingLot) (*ParkingLot, common.AppError) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.l.Error("error creating parking lot transaction", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				r.l.Error("error rolling back the transaction", "err", txErr)
			}
		}
	}()

	if appErr := r.parkingLotExistsByName(ctx, tx, lot.Name); appErr != nil {
		return nil, appErr
	}

	var plUUID uuid.UUID
	var plID int
	err = tx.QueryRowContext(ctx, "INSERT INTO parking_lots (name) VALUES ($1) RETURNING id, uuid;", lot.Name).Scan(&plID, &plUUID)
	if err != nil {
		r.l.Error("error creating parking lot", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	slots, csErr := r.createSlots(ctx, tx, plID, lot.DesiredSlots)
	if err != nil {
		return nil, csErr
	}

	if err := tx.Commit(); err != nil {
		r.l.Error("error committing transaction", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	lot.Slots = slots
	lot.ID = plUUID
	return lot, nil
}

// parkingLotExistsByName determines if a parking lot with the given name exists, used to prevent duplicate names.
// ToDO: Think about creating a database index on name column.
func (r *ParkingLotRepoDB) parkingLotExistsByName(ctx context.Context, tx *sql.Tx, name string) common.AppError {
	var exists bool
	err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM parking_lots WHERE name = $1)", name).Scan(&exists)
	if err != nil {
		r.l.Error("error checking parking lot existence", "err", err)
		return common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if exists {
		r.l.Error("parking lot with this name already exists", "name", name)
		return common.NewConflictError("parking lot with this name already exists")
	}

	return nil
}

// createSlots inserts bulk amount of slots and returns error if exists.
func (r *ParkingLotRepoDB) createSlots(ctx context.Context, tx *sql.Tx, lotID int, numSlots int) ([]Slot, common.AppError) {
	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO slots (parking_lot_id, slot_number) 
        VALUES ($1, $2)
        RETURNING uuid
    `)
	if err != nil {
		r.l.Error("error preparing slot creation statement", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	defer stmt.Close()

	createdSlots := make([]Slot, 0, numSlots)
	for i := 1; i <= numSlots; i++ {
		var slotUUID uuid.UUID
		execErr := stmt.QueryRowContext(ctx, lotID, i).Scan(&slotUUID)
		if execErr != nil {
			r.l.Error("error creating slots", "err", execErr)
			return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, execErr)
		}

		createdSlots = append(createdSlots, Slot{
			ID:            slotUUID,
			SlotNumber:    i,
			IsAvailable:   true,
			IsMaintenance: false,
		})
	}

	return createdSlots, nil
}
