// internal/web/server.go
// Defines the backend that uses the filesystem for content, sqlite fo metadata
package web

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/steadyloop/truetone/internal/transcode"
)

type server struct {
	tempDir  string
	metadata MetadataStore
	content  ContentStore
	mux      *http.ServeMux
}

func NewServer(metadata MetadataStore, content ContentStore, tempDir string) *server {
	return &server{metadata: metadata, content: content, tempDir: tempDir}
}

func (s *server) Start(lis net.Listener) error {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/watch/{id}", s.handleWatch)
	s.mux.HandleFunc("/content/{id}/{filename}", s.handleContent)
	return http.Serve(lis, s.mux)
}

// DONE
// GET / - Renders the home screen (index), with []content from metadata.List()
func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	videos, err := s.metadata.List()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// Reuses the metadata struct and parses into html; html must reformat the time as text.
	if err := tmpl.ExecuteTemplate(w, "library.html", videos); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// POST /upload - Uploads requests in temp_dir -> ffmpeg -> HLS -> serve.
// 400 if field is missing, 409 if id exists. 303 else.
func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field missing", http.StatusBadRequest) // 400
		return
	}
	defer file.Close()
	// 2. Create tempDir
	id := uuid.NewString()
	tempDir := filepath.Join(s.tempDir, id)
	if err := os.Mkdir(tempDir, 0755); err != nil {
		http.Error(w, "UUID collision!!", http.StatusInternalServerError)
	} // fails loud if it somehow already exists
	// 3. Upload
	video := filepath.Join(tempDir, header.Filename)
	f, err := os.Create(video)
	if err != nil {
		http.Error(w, "error creating file on server", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(f, file); err != nil {
		f.Close()
		http.Error(w, "error transfering file to server", http.StatusInternalServerError)
		return
	}
	f.Close()
	// 4. Process the upload
	s.metadata.Create(id, header.Filename, time.Now())
	go s.processUpload(id, tempDir)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) processUpload(id, tempDir string) {
	// Go into the tempfile, run the transcoding, and update the metadata as you go.
	// 1.
	defer os.RemoveAll(tempDir)
	transcode.ToHLS(tempDir, tempDir)
	s.content.Write(id, tempDir)

	// defer os.RemoveAll(tempDir) as its first line, then:
	// mkdir hls/ → transcode.ToHLS → loop hls/ into s.content.Write
	// → transcode.Duration (fallback 0 on error) → s.metadata.MarkReady. Every failure: log.Printf + MarkFailed.

	// MarkFailed(id string) error — add to interfaces.go, implement in sqlite.go.
	// Duration(path string) (int, error) — add to transcode.go (only ToHLS exists there right now).

}

// Done
// GET /watch/{id} — render "watch.html" with the video's metadata. 404 if it doesn't exist.
func (s *server) handleWatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	log.Println("Streaming: ", id)
	meta, err := s.metadata.Read(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if meta == nil {
		http.NotFound(w, r)
		return
	}
	// Renders "watch.html" with the metadata corresponding to the id specified by the request.
	if err := tmpl.ExecuteTemplate(w, "watch.html", meta); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// DONE
// GET /content/{id}/{filename} — returns the HLS binary, 500 on failure
func (s *server) handleContent(w http.ResponseWriter, r *http.Request) {
	id, filename := r.PathValue("id"), r.PathValue("filename")
	data, err := s.content.Read(id, filename)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError) // 500
		return
	}
	w.Write(data)
}
