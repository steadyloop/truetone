// SQLiteMetadataStore: Stores the video metadata in single sqlite table
package web

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// implementing service in internal/web/interfaces.go
type SQLiteMetadataStore struct {
	db *sql.DB
}

// Ensures SQLiteMetadataStore implements MetadataStore
var _ MetadataStore = (*SQLiteMetadataStore)(nil)

// ┌──────────────────────────────────┐
// │              videos              │
// ├──────────────────────────────────┤
// │ id		      TEXT      PK        │
// | title        TEXT      NOT NULL  |
// │ state        TEXT      NOT NULL  │
// │ seconds      INT                 │
// │ uploaded_at  DATETIME  NOT NULL  │
// └──────────────────────────────────┘
func NewSQLiteMetadataStore(path string) (*SQLiteMetadataStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	query := `
		CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		state TEXT NOT NULL,
		seconds INT,
		uploaded_at DATETIME NOT NULL	
		);`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return &SQLiteMetadataStore{db: db}, nil
}

// Create saves a new entry. A duplicate video ID must return an error (handler turns it into 409).
func (s *SQLiteMetadataStore) Create(id, title string, uploadedAt time.Time) (err error) {
	_, err = s.db.Exec("INSERT INTO videos (id, title, uploaded_at, state) VALUES (?, ?, ?, ?)", id, title, uploadedAt.UTC(), "processing")
	if err != nil {
		return fmt.Errorf("failed to insert video: %w", err)
	}
	return
}

// Lists all video entries (newest first)
func (s *SQLiteMetadataStore) List() ([]Metadata, error) {
	// 1. Retreive pointer to start querying
	rows, err := s.db.Query("SELECT id, title, state, seconds, uploaded_at FROM videos ORDER BY uploaded_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 2. Save into the struct as query proceeds
	var res []Metadata
	for rows.Next() {
		var entry Metadata
		if err := rows.Scan(&entry.ID, &entry.Title, &entry.State, &entry.DurationS, &entry.UploadedAt); err != nil {
			return nil, err
		}
		res = append(res, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// Read returns the entry for videoId, or nil (no error) if it doesn't exist.
func (s *SQLiteMetadataStore) Read(id string) (*Metadata, error) {
	var meta Metadata
	row := s.db.QueryRow("SELECT id, title, state, seconds, uploaded_at FROM videos WHERE id = ?", id)
	switch err := row.Scan(&meta.ID, &meta.Title, &meta.State, &meta.DurationS, &meta.UploadedAt); {
	case err == nil:
		return &meta, nil
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	default:
		return nil, err
	}
}

// MarkReady transitions a video from "processing" to "ready" once transcoding
// finishes, recording its duration.
func (s *SQLiteMetadataStore) MarkReady(id string, seconds int) error {
	_, err := s.db.Exec("UPDATE videos SET state = ?, seconds = ? WHERE id = ?", "ready", seconds, id)
	if err != nil {
		return fmt.Errorf("failed to mark video ready: %w", err)
	}
	return nil
}
