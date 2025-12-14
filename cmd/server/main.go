package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/colors"
	"github.com/gary-norman/forum/internal/http/routes"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/view"
	"github.com/gary-norman/forum/internal/workers"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Colors, _ = colors.UseFlavor("Mocha")
	ErrorMsgs = models.CreateErrorMessages()
)

func main() {
	// Initialize app first so it's available to everything
	appInstance, cleanup, err := app.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// CLI commands: migrate + seed
	if len(os.Args) > 1 {
		defer cleanup() // Cleanup after CLI commands
		switch os.Args[1] {
		case "migrate":
			migrations, err := discoverMigrations("./migrations")
			if err != nil {
				log.Fatalf("Failed to discover migrations: %v", err)
			}
			if err := runMigrations(appInstance.DB, migrations); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
			return
		case "seed":
			if err := runSeed(appInstance.DB); err != nil {
				log.Fatalf("Seeding failed: %v", err)
			}
			return
		}
	}

	// Auto-run migrations in dev mode
	if os.Getenv("DB_ENV") == "dev" {
		migrations, err := discoverMigrations("./migrations")
		if err != nil {
			log.Printf("Warning: Failed to discover migrations: %v", err)
		} else {
			if err := runMigrations(appInstance.DB, migrations); err != nil {
				log.Printf("Warning: Auto-migration failed: %v", err)
			}
		}
	}

	// Otherwise, start the web server (pass cleanup for graceful shutdown)
	startServer(appInstance, cleanup)
}

func startServer(appInstance *app.App, dbCleanup func()) {
	// pprof server for profiling
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Init template helper
	th := view.TempHelper{App: appInstance}
	th.Init()

	// Create and start logger pool
	loggerPool := workers.NewLoggerPool(3, 1000, appInstance.DB)
	loggerPool.Start()

	// Router
	router := routes.NewRouter(appInstance, loggerPool)

	port := 8888
	portStr := fmt.Sprintf(Colors.CodexPink+"%d"+Colors.Reset, port)
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	address := "port: " + Colors.CodexPink + portStr + Colors.Reset
	fmt.Printf(ErrorMsgs.KeyValuePair, "Starting server on port", port)
	log.Printf(ErrorMsgs.ConnSuccess, address)

	// Run server
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf(ErrorMsgs.ConnInit, srv.Addr, "srv.ListenAndServe")
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println(Colors.Teal + "Stopped serving new connections." + Colors.Reset)
	}()

	// Graceful shutdown
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)
	<-shutdownSignal

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	log.Println("Shutting down gracefully...")

	// 1. Shutdown HTTP server (stops accepting new requests)
	log.Println("Stopping HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf(ErrorMsgs.Shutdown, err)
	}

	// 2. Shutdown logger pool (drain pending logs)
	log.Println("Draining log queue...")
	if err := loggerPool.Shutdown(shutdownCtx); err != nil {
		log.Printf("Warning: Logger pool shutdown timeout: %v", err)
	}

	// 3. Wait for in-flight database operations to complete
	// srv.Shutdown() already waits for handlers to complete, which includes their DB queries
	// This ensures all active queries finish before we close the DB
	log.Println("Waiting for database operations to complete...")

	// 4. Close database connection
	log.Println("Closing database connection...")
	dbCleanup()

	log.Println(Colors.Teal + "Graceful shutdown complete." + Colors.Reset)
}
