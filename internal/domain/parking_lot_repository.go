package domain

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/google/uuid"
)

// ParkingLotRepository defines the interface for interacting with parking lot data in the postgresql database.
type ParkingLotRepository interface {
	CreateParkingLot(ctx context.Context, lot *ParkingLot) (*ParkingLot, common.AppError)

	// GetParkingLotByID(ctx context.Context, id uuid.UUID) (*ParkingLot, common.AppError)
	// GetParkingLotSlots(ctx context.Context, id uuid.UUID) (Slot, common.AppError)
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

// CreateParkingLot Creates a parking lot. If a parking lot with the same name exists, returns a 409 Conflict error.
// Otherwise, returns a 500 Internal Server Error on failure.
func (r *ParkingLotRepoDB) CreateParkingLot(ctx context.Context, lot *ParkingLot) (*ParkingLot, common.AppError) {
	if appErr := r.parkingLotExistsByName(ctx, lot.Name); appErr != nil {
		return nil, appErr
	}

	parkingLotInsertQuery := `INSERT INTO parking_lots (name) VALUES ($1) RETURNING parking_lot_id;`

	var createdID uuid.UUID
	err := r.db.QueryRowContext(ctx, parkingLotInsertQuery, lot.Name).Scan(&createdID)
	if err != nil {
		r.l.Error("error inserting parking lot", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	lot.ID = createdID
	return lot, nil
}

// parkingLotExistsByName determines if a parking lot with the given name exists, used to prevent duplicate names.
func (r *ParkingLotRepoDB) parkingLotExistsByName(ctx context.Context, name string) common.AppError {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM parking_lots WHERE name = $1)", name).Scan(&exists)
	if err != nil {
		r.l.Error("error checking parking lot existence", err, err.Error())
		return common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	if exists {
		r.l.Error("db field conflict error", "err", errors.New("parking lot with this name already exists"))
		return common.NewConflictError("parking lot with this name already exists")
	}

	return nil
}
