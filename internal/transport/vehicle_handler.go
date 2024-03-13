package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ashtishad/gopark/internal/domain"
	"github.com/google/uuid"
)

// ParkVehicleRequest represents the information needed to park a vehicle in the HTTP request body
type ParkVehicleRequest struct {
	ParkingLotID       string `json:"parkingLotId"`
	RegistrationNumber string `json:"registrationNumber"`
}

// UnparkVehicleRequest represents the request for unparking
type UnparkVehicleRequest struct {
	RegistrationNumber string `json:"registrationNumber"`
}

type VehicleHandler struct {
	Repo   *domain.VehicleRepositoryDB
	Logger *slog.Logger
}

// Park handles HTTP requests to park a vehicle
func (h *VehicleHandler) Park(w http.ResponseWriter, r *http.Request) {
	var reqBody ParkVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		return
	}

	parkingLotID, err := uuid.Parse(r.PathValue("id")) // go 1.22 introduced path param from routes.
	if err != nil {
		writeResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid parking lot ID format"})
		return
	}

	parkedVehicle, appErr := h.Repo.ParkVehicle(r.Context(), parkingLotID, reqBody.RegistrationNumber)
	if appErr != nil {
		writeResponse(w, appErr.Code(), map[string]string{"error": appErr.Error()})
		return
	}

	writeResponse(w, http.StatusOK, parkedVehicle)
}

// Unpark handles HTTP requests for unparking vehicles
func (h *VehicleHandler) Unpark(w http.ResponseWriter, r *http.Request) {
	var reqBody UnparkVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		return
	}

	unparkedVehicle, appErr := h.Repo.UnparkVehicle(r.Context(), reqBody.RegistrationNumber)
	if appErr != nil {
		writeResponse(w, appErr.Code(), map[string]string{"error": appErr.Error()})
		return
	}

	writeResponse(w, http.StatusOK, unparkedVehicle)
}
