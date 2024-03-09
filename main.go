package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/ashtishad/gopark/db/conn"
	"github.com/ashtishad/gopark/internal/common"
)

func main() {
	handlerOpts := common.GetSlogConf()
	l := slog.New(slog.NewTextHandler(os.Stdout, handlerOpts))
	slog.SetDefault(l)

	dbClient := conn.GetDBClient(l)

	defer func(dbClient *sql.DB) {
		if dbClsErr := dbClient.Close(); dbClsErr != nil {
			l.Error("unable to close db", "err", dbClsErr)
			os.Exit(1)
		}
	}(dbClient)

	fmt.Println("Hello from GoPark")
}
