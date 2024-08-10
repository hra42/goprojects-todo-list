package duckdb

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

var (
	db   *sql.DB
	once sync.Once
)

// Task represents a task in our application
type Task struct {
	ID          int
	Description string
	CreatedAt   time.Time
	IsComplete  bool
}

// InitDB initializes the DuckDB database
func InitDB(dbPath string) error {
	var err error
	once.Do(func() {
		db, err = sql.Open("duckdb", dbPath)
		if err != nil {
			log.Printf("Failed to open database: %v", err)
			return
		}

		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping database: %v", err)
			return
		}

		err = createTable()
		if err != nil {
			log.Printf("Failed to create table: %v", err)
			return
		}
	})
	return err
}

// createTable creates the tasks table and ID_SEQ sequence if they don't exist
func createTable() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Try to create the sequence. If it already exists, this will fail, but that's okay.
	_, _ = tx.Exec("CREATE SEQUENCE IF NOT EXISTS ID_SEQ START 1")

	// Create the table if it doesn't exist
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY DEFAULT NEXTVAL('ID_SEQ'),
			description TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			is_complete BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AddTask adds a new task to the database
func AddTask(description string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := db.Exec("INSERT INTO tasks (description, created_at) VALUES (?, ?)",
		description, time.Now())
	return err
}

// ListTasks returns all tasks, optionally filtered by completion status
func ListTasks(showAll bool) ([]Task, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	query := "SELECT id, description, created_at, is_complete FROM tasks"
	if !showAll {
		query += " WHERE is_complete = FALSE"
	}
	query += " ORDER BY id ASC"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Description, &t.CreatedAt, &t.IsComplete)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// CompleteTask marks a task as complete
func CompleteTask(id int) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := db.Exec("UPDATE tasks SET is_complete = TRUE WHERE id = ?", id)
	return err
}

// DeleteTask deletes a task from the database
func DeleteTask(id int) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// CloseDB closes the database connection
func CloseDB() {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}
