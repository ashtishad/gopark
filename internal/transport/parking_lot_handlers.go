package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/ashtishad/gopark/internal/domain"
)

type ParkingLotHandler struct {
	Repo   *domain.ParkingLotRepoDB
	Logger *slog.Logger
}

func (h *ParkingLotHandler) CreateParkingLot(w http.ResponseWriter, r *http.Request) {
	var newLot domain.ParkingLot
	if err := json.NewDecoder(r.Body).Decode(&newLot); err != nil {
		http.Error(w, common.NewBadRequestError("invalid request payload").Error(), http.StatusBadRequest)
		return
	}

	// Simplified, use regex for comprehensive Input validation.
	if newLot.Name == "" {
		http.Error(w, common.NewBadRequestError("parking lot name is required").Error(), http.StatusBadRequest)
		return
	}

	createdLot, appErr := h.Repo.CreateParkingLot(r.Context(), &newLot)
	if appErr != nil {
		writeResponse(w, appErr.Code(), map[string]string{"error": appErr.Error()})
		return
	}

	writeResponse(w, http.StatusCreated, createdLot)
}
