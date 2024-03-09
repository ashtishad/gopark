package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ashtishad/gopark/internal/common"
	"github.com/ashtishad/gopark/internal/domain"
	"github.com/google/uuid"
)

// ParkVehicleRequest represents the information needed to park a vehicle in the HTTP request body
type ParkVehicleRequest struct {
	ParkingLotID       string `json:"parkingLotId"`
	RegistrationNumber string `json:"registrationNumber"`
}

// ParkVehicleResponse represents the parked vehicle information returned in the HTTP response
type ParkVehicleResponse struct {
	Vehicle *domain.Vehicle `json:"vehicle"`
}

type VehicleHandler struct {
	Repo   *domain.VehicleRepositoryDB
	Logger *slog.Logger
}

// Park handles HTTP requests to park a vehicle
func (h *VehicleHandler) Park(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, common.NewBadRequestError("only POST requests supported").Error(), http.StatusBadRequest)
		return
	}

	var requestBody ParkVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, common.NewBadRequestError("invalid request payload").Error(), http.StatusBadRequest)
		return
	}

	parkingLotID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, common.NewBadRequestError("invalid parking lot ID format").Error(), http.StatusBadRequest)
		return
	}

	parkedVehicle, appErr := h.Repo.ParkVehicle(r.Context(), parkingLotID, requestBody.RegistrationNumber)
	if appErr != nil {
		writeResponse(w, appErr.Code(), map[string]string{"error": appErr.Error()})
		return
	}

	writeResponse(w, http.StatusCreated, parkedVehicle)
}
