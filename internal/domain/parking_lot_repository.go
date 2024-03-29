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

// ParkingLotRepository defines the interface for interacting with parking lot data in the postgresql database.
type ParkingLotRepository interface {
	CreateParkingLot(ctx context.Context, lot *ParkingLot) (*ParkingLot, common.AppError)
	GetParkingLotStatus(ctx context.Context, plUUID uuid.UUID) (*ParkingLotStatus, common.AppError)
	GetDailyReport(ctx context.Context, parkingLotID uuid.UUID, dateString string) (*DailyReport, common.AppError)
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
		r.l.Error(common.ErrTXBegin, "err", err, "src", "CreateParkingLot")
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				r.l.Error(common.ErrTXRollback, "err", txErr, "src", "CreateParkingLot")
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

	if cmtErr := tx.Commit(); cmtErr != nil {
		r.l.Error(common.ErrTxCommit, "err", err, "src", "CreateParkingLot")
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, cmtErr)
	}

	lot.Slots = slots
	lot.ID = plUUID
	return lot, nil
}

// parkingLotExistsByName determines if a parking lot with the given name exists, used to prevent duplicate names.
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

// GetParkingLotStatus retrieves the current status of a parking lot, including the name of the
// parking lot and the status of each slot. This information is essential for parking managers
// to monitor occupancy and identify available parking spaces, returns errors if exists.
func (r *ParkingLotRepoDB) GetParkingLotStatus(ctx context.Context, plUUID uuid.UUID) (*ParkingLotStatus, common.AppError) {
	plID, apiErr := getIDByUUID(ctx, r.db, r.l, tableParkingLots, plUUID)
	if apiErr != nil {
		return nil, apiErr
	}

	var parkingLotName string
	var slots []SlotStatus

	if err := r.db.QueryRowContext(ctx, `
        SELECT name FROM parking_lots WHERE id = $1`, plID).Scan(&parkingLotName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFoundError("parking lot not found")
		} else if err != nil {
			r.l.Error("unable to get parking lot name", "err", err)
			return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
		}
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT s.uuid, v.registration_number, v.parked_at, v.unparked_at
        FROM slots s
        LEFT JOIN vehicles v ON v.slot_id = s.id
        WHERE s.parking_lot_id = $1`, plID)
	if err != nil {
		r.l.Error("unable to get slot info", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}
	defer rows.Close()

	for rows.Next() {
		var slot SlotStatus
		if scnErr := rows.Scan(&slot.SlotID, &slot.RegistrationNum, &slot.ParkedAt, &slot.UnparkedAt); scnErr != nil {
			r.l.Error("unable to scan slot info", "err", scnErr)
			return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, scnErr)
		}

		slots = append(slots, slot)
	}

	return &ParkingLotStatus{
		ParkingLotID: plUUID,
		Name:         parkingLotName,
		Slots:        slots,
	}, nil
}

// GetDailyReport generates a report summarizing parking activity for a specific parking lot on a given date.
// This includes the total number of vehicles parked, total parking hours (rounded up), and total fees collected.
// The report is essential for parking lot managers to analyze usage and revenue.
// Query Explanation:
// 1. Calculates the total vehicles parked using COUNT(*).
// 2. Calculates total parking hours by summing durations (in seconds) after applying CEIL to round up to the nearest hour.
// 3. Calculates total fees by multiplying the rounded parking hours with the hourly rate (10).
func (r *ParkingLotRepoDB) GetDailyReport(ctx context.Context, plUUID uuid.UUID, reportDate time.Time) (*DailyReport, common.AppError) {
	startDate := reportDate
	endDate := reportDate.AddDate(0, 0, 1)

	plID, appErr := getIDByUUID(ctx, r.db, r.l, tableParkingLots, plUUID)
	if appErr != nil {
		return nil, appErr
	}

	var report DailyReport
	sqlDailyReport := `
   SELECT 
       COUNT(*) as total_vehicles_parked,
       SUM(CEIL(EXTRACT(EPOCH FROM (v.unparked_at - v.parked_at))/3600)) as total_parking_hours, -- Ceil parking duration
       SUM(CEIL(EXTRACT(EPOCH FROM (v.unparked_at - v.parked_at))/3600) * 10)::int as total_fee_collected
   FROM vehicles v
   JOIN slots s ON v.slot_id = s.id
   JOIN parking_lots pl ON s.parking_lot_id = pl.id
   WHERE pl.id = $1 
    AND v.parked_at >= $2 AND v.parked_at < $3 
`
	err := r.db.QueryRowContext(ctx, sqlDailyReport, plID, startDate, endDate).Scan(
		&report.TotalVehiclesParked,
		&report.TotalParkingHours,
		&report.TotalFeeCollected)

	if errors.Is(err, sql.ErrNoRows) {
		r.l.Error("no records found for this date or parking lot", "err", err)
		return nil, common.NewNotFoundError("no records found for this date or parking lot")
	} else if err != nil {
		r.l.Error("error generating daily report", "err", err)
		return nil, common.NewInternalServerError(common.ErrUnexpectedDatabase, err)
	}

	return &report, nil
}
