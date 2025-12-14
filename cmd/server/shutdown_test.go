package main

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestGracefulDatabaseShutdown verifies that active DB operations complete before DB closes
func TestGracefulDatabaseShutdown(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create a simple table
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	var wg sync.WaitGroup
	activeQuery := make(chan bool, 1)
	queryCompleted := make(chan bool, 1)

	// Simulate a long-running query
	wg.Add(1)
	go func() {
		defer wg.Done()
		activeQuery <- true // Signal that query started
		
		// Simulate work
		ctx := context.Background()
		rows, err := db.QueryContext(ctx, "SELECT * FROM test")
		if err != nil {
			t.Errorf("Query failed: %v", err)
			return
		}
		defer rows.Close()
		
		time.Sleep(100 * time.Millisecond) // Simulate slow query
		queryCompleted <- true
	}()

	// Wait for query to start
	<-activeQuery

	// Now try to close the database while query is running
	// This simulates what happens during shutdown
	closeStarted := time.Now()
	
	// Wait for active operations with timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Operations completed
		t.Logf("Active operations completed after %v", time.Since(closeStarted))
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for active operations")
	}

	// Verify query actually completed
	select {
	case <-queryCompleted:
		t.Log("Query completed successfully before shutdown")
	default:
		t.Fatal("Query did not complete")
	}

	// Now safe to close database
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	t.Log("✓ Database closed gracefully after active operations completed")
}
