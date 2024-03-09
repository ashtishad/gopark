package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ashtishad/gopark/db/conn"
	"github.com/ashtishad/gopark/internal/common"
	"github.com/ashtishad/gopark/internal/domain"
	"github.com/ashtishad/gopark/internal/transport"
)

func main() {
	// 1. Initialize structured logger
	handlerOpts := common.GetSlogConf()
	logger := slog.New(slog.NewTextHandler(os.Stdout, handlerOpts))
	slog.SetDefault(logger)

	// 2. Check environment variables, if not exists sets default.
	sanityCheck(logger)

	// 3. Get postgres database client
	dbClient := conn.GetDBClient(logger)

	defer func(dbClient *sql.DB) {
		if dbClsErr := dbClient.Close(); dbClsErr != nil {
			logger.Error("unable to close db", "err", dbClsErr)
			os.Exit(1)
		}
	}(dbClient)

	// 4. Wire up dependencies
	parkingLotRepo := domain.NewParkingLotRepoDB(dbClient, logger)
	parkingLotHandler := transport.ParkingLotHandler{Repo: parkingLotRepo, Logger: logger}

	// 5. Structured Server Configuration
	srv := &http.Server{
		Addr:              net.JoinHostPort(os.Getenv("API_HOST"), os.Getenv("API_PORT")),
		Handler:           nil,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// 6. Route Registration (using a router or a simple mux)
	router := http.NewServeMux()
	router.HandleFunc("POST /parking-lots", parkingLotHandler.CreateParkingLot)
	srv.Handler = router

	// 7. Start the Server
	logger.Info("Server starting...", slog.String("address", srv.Addr))
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Error starting server", "err", err)
		os.Exit(1)
	}
}

// sanityCheck checks essential env variables required ot run the app, sets defaults if not exists
func sanityCheck(l *slog.Logger) {
	defaultEnvVars := map[string]string{
		"API_HOST":  "127.0.0.1",
		"API_PORT":  "8000",
		"DB_USER":   "postgres",
		"DB_PASSWD": "postgres",
		"DB_HOST":   "127.0.0.1",
		"DB_PORT":   "5432",
		"DB_NAME":   "gopark",
	}

	for key, defaultValue := range defaultEnvVars {
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, defaultValue); err != nil {
				l.Error(fmt.Sprintf(
					"failed to set environment variable %s to default value %s. Exiting application.",
					key,
					defaultValue,
				))
				os.Exit(1)
			}

			l.Warn(fmt.Sprintf("environment variable %s not defined. Setting to default: %s", key, defaultValue))
		}
	}
}
