// Command example starts a restinpieces application that serves its own
// metrics: the default recorder counts requests and request time, and the
// internal daemon serves them on metrics.listen_addr.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/caasmo/restinpieces"
	"github.com/caasmo/restinpieces-metrics"
)

func main() {
	dbPath := flag.String("dbpath", "", "Path to the SQLite database file (required)")
	ageKeyPath := flag.String("agekey", "", "Path to the age identity (private key) file (required)")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s -dbpath <database-path> -agekey <identity-file-path>\n\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "Start the restinpieces application server with the default metrics.\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *dbPath == "" || *ageKeyPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	dbPool, err := restinpieces.NewModerncPool(*dbPath)
	if err != nil {
		slog.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}

	defer func() {
		slog.Info("Closing database pool...")
		if err := dbPool.Close(); err != nil {
			slog.Error("Error closing database pool", "error", err)
		}
	}()

	app, srv, err := restinpieces.New(
		restinpieces.WithModerncPool(dbPool),
		restinpieces.WithAgeKeyPath(*ageKeyPath),
	)
	if err != nil {
		slog.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	recorder, daemon := metrics.NewDefault(app.ConfigPointer(), app.Logger())
	app.SetMetrics(recorder)
	srv.AddDaemon(daemon)

	srv.Run()

	slog.Info("Server shut down gracefully.")
}
