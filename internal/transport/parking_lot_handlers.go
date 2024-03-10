package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ashtishad/gopark/internal/domain"
	"github.com/google/uuid"
)

type ParkingLotHandler struct {
	Repo   *domain.ParkingLotRepoDB
	Logger *slog.Logger
}

func (h *ParkingLotHandler) CreateParkingLot(w http.ResponseWriter, r *http.Request) {
	var newLot domain.ParkingLot
	if err := json.NewDecoder(r.Body).Decode(&newLot); err != nil {
		writeResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		return
	}

	// Simplified, use regex for comprehensive Input validation.
	if newLot.Name == "" {
		writeResponse(w, http.StatusBadRequest, map[string]string{"error": "parking lot name is required"})
		return
	}

	createdLot, appErr := h.Repo.CreateParkingLot(r.Context(), &newLot)
	if appErr != nil {
		writeResponse(w, appErr.Code(), map[string]string{"error": appErr.Error()})
		return
	}

	writeResponse(w, http.StatusCreated, createdLot)
}

func (h *ParkingLotHandler) GetParkingLotStatus(w http.ResponseWriter, r *http.Request) {
	plUUID, err := uuid.Parse(r.PathValue("id")) // go 1.22 introduced path param from routes.
	if err != nil {
		writeResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid parking lot ID format"})
		return
	}

	status, appErr := h.Repo.GetParkingLotStatus(r.Context(), plUUID)
	if appErr != nil {
		writeResponse(w, appErr.Code(), map[string]string{"error": appErr.Error()})
		return
	}

	writeResponse(w, http.StatusOK, status)
}
