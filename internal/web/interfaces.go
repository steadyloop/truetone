// internal/web/interfaces.go
// Contains schemes that internal/web/sqlite.go uses

package web

import "time"

type State string

const (
	Processing State = "processing"
	Ready      State = "ready"
	Failed     State = "failed"
)

type Metadata struct {
	ID         string
	State      State
	Title      string
	DurationS  int
	UploadedAt time.Time
}

type MetadataStore interface {
	Create(id, title string, uploadedAt time.Time) error
	List() ([]Metadata, error)
	Read(id string) (*Metadata, error)
	MarkReady(id string, seconds int) error
	MarkFailed(id string) error
}

type ContentStore interface {
	Read(id string, filename string) ([]byte, error)
	Write(id string, filename string, data []byte) error
}
