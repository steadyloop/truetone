// internal/web/interfaces.go
// Contains schemes that internal/web/sqlite.go uses

package web

import "time"

type Metadata struct {
	Id         string
	Title      string
	UploadedAt time.Time
}

type MetadataStore interface {
	Read(id string) (*Metadata, error)
	List() ([]Metadata, error)
	Create(id, title string, uploadedAt time.Time) error
}

type ContentStore interface {
	Read(id string, filename string) ([]byte, error)
	Write(id string, filename string, data []byte) error
}
