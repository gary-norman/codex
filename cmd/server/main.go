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
	defer cleanup()

	// CLI commands: migrate + seed
	if len(os.Args) > 1 {
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

	// Otherwise, start the web server
	startServer(appInstance)
}

func startServer(appInstance *app.App) {
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

	// TODO(human): Exercise 5 - Implement graceful shutdown sequence
	//
	// Current problem: Database might close while queries are still running!
	// The defer cleanup() in main() will close the DB as soon as main() exits,
	// but handlers might still be processing requests with active DB queries.
	//
	// Your task: Implement proper shutdown order
	// 1. Shutdown HTTP server (srv.Shutdown stops accepting new requests, waits for handlers)
	// 2. Shutdown logger pool (drain pending logs with loggerPool.Shutdown)
	// 3. Add log message explaining that srv.Shutdown already waited for DB operations
	// 4. Close database connection (hint: you'll need to pass cleanup function to startServer)
	//
	// Hints:
	// - srv.Shutdown(shutdownCtx) blocks until all handlers complete
	// - Since handlers run DB queries synchronously, waiting for handlers = waiting for queries
	// - You'll need to modify the function signature: startServer(appInstance *app.App, dbCleanup func())
	// - Call dbCleanup() after everything else shuts down
	// - Move defer cleanup() from main() to only run for CLI commands

	// Your implementation here:
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf(ErrorMsgs.Shutdown, err)
	}

	log.Println("Draining log queue...")
	if err := loggerPool.Shutdown(shutdownCtx); err != nil {
		log.Printf("Warning: Logger pool shutdown timeout: %v", err)
	}

	log.Println(Colors.Teal + "Graceful shutdown complete." + Colors.Reset)
}
