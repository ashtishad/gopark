package domain

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/google/uuid"
)

// getParkingLotIdByUUID retrieves the internal integer ID of a parking lot given its UUID.
func getParkingLotIDByUUID(ctx context.Context, db *sql.DB, l *slog.Logger, parkingLotUUID uuid.UUID) (int, common.AppError) {
	var parkingLotID int

	err := db.QueryRowContext(ctx, `SELECT id FROM parking_lots WHERE uuid = $1`, parkingLotUUID).Scan(&parkingLotID)

	if errors.Is(err, sql.ErrNoRows) {
		l.Error("parking lot not found", "err", err)
		return 0, common.NewNotFoundError(common.ErrUnexpectedDatabase)
	} else if err != nil {
		l.Error("error fetching parking lot ID by uuid", "err", err)
		return 0, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return parkingLotID, nil
}

// getSlotIDByUUID retrieves the internal integer ID of a slot given its UUID.
func getSlotIDByUUID(ctx context.Context, db *sql.DB, l *slog.Logger, slotUUID uuid.UUID) (int, common.AppError) {
	var slotID int

	err := db.QueryRowContext(ctx, `SELECT id FROM slots WHERE uuid = $1`, slotUUID).Scan(&slotID)

	if errors.Is(err, sql.ErrNoRows) {
		l.Error("slot not found", "err", err)
		return 0, common.NewNotFoundError(common.ErrUnexpectedDatabase)
	} else if err != nil {
		l.Error("error fetching slot ID by uuid", "err", err)
		return 0, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return slotID, nil
}

// getVehicleIdByUUID retrieves the internal integer ID of a vehicle given its UUID.
func getVehicleIDByUUID(ctx context.Context, db *sql.DB, l *slog.Logger, vehicleUUID uuid.UUID) (int, common.AppError) {
	var vehicleID int

	err := db.QueryRowContext(ctx, `SELECT id FROM vehicles WHERE uuid = $1`, vehicleUUID).Scan(&vehicleID)

	if errors.Is(err, sql.ErrNoRows) {
		l.Error("vehicle not found", "err", err)
		return 0, common.NewNotFoundError(common.ErrUnexpectedDatabase)
	} else if err != nil {
		l.Error("error fetching vehicle ID by uuid", "err", err)
		return 0, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return vehicleID, nil
}
