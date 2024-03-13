package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/google/uuid"
)

const (
	tableParkingLots = "parking_lots"
	tableSlots       = "slots"
	tableVehicles    = "vehicles"
)

func getIDByUUID(ctx context.Context, db *sql.DB, l *slog.Logger, tableName string, uuid uuid.UUID) (int, common.AppError) {
	var id int

	query := fmt.Sprintf("SELECT id FROM %s WHERE uuid = $1", tableName) //nolint:gosec // this is internal method
	err := db.QueryRowContext(ctx, query, uuid).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		l.Error(fmt.Sprintf("%s not found", tableName), "err", err)
		return 0, common.NewNotFoundError(common.ErrUnexpectedDatabase)
	} else if err != nil {
		l.Error(fmt.Sprintf("error fetching %s ID by uuid", tableName), "err", err)
		return 0, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return id, nil
}
