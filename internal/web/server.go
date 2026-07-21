// internal/web/server.go
// Defines the backend that uses the filesystem for content, sqlite fo metadata
package web

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/steadyloop/truetone/internal/transcode"
)

const uploadFile = "upload.mp4" // file names are unified, nested under the video id path

type server struct {
	jobs     string
	metadata MetadataStore
	content  ContentStore
	mux      *http.ServeMux
}

func NewServer(metadata MetadataStore, content ContentStore, jobPath string) *server {
	return &server{metadata: metadata, content: content, jobs: jobPath}
}

func (s *server) Start(lis net.Listener) error {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("GET /upload", s.handleUploadPage)
	s.mux.HandleFunc("POST /upload", s.handleUpload)
	s.mux.HandleFunc("/watch/{id}", s.handleWatch)
	s.mux.HandleFunc("/content/{id}/{filename}", s.handleContent)
	s.mux.Handle("/static/", http.FileServerFS(staticFS))
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

// GET /upload — renders the dedicated upload page.
func (s *server) handleUploadPage(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "upload.html", nil); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// POST /upload - Uploads requests in job -> ffmpeg -> HLS -> serve.
// 400 if field is missing, 409 if id exists. 303 else.
func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field missing", http.StatusBadRequest) // 400
		return
	}
	defer file.Close()
	// 2. Create the job path
	id := uuid.NewString()
	slog.Info("upload started", "id", id, "filename", header.Filename)
	job := filepath.Join(s.jobs, id)
	if err := os.Mkdir(job, 0755); err != nil {
		http.Error(w, "UUID collision!!", http.StatusInternalServerError)
		return
	} // fails loud if it somehow already exists
	// 3. Upload
	videoPath := filepath.Join(job, uploadFile)
	f, err := os.Create(videoPath)
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
	// 4. Process the upload by adding entry into metadata store. Actual upload handled by processUpload() in background
	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
	}
	if err := s.metadata.Create(id, title, time.Now()); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	go s.processUpload(id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// processUpload runs in the background after handleUpload returns: input.mp4 ->
// ffmpeg -> HLS -> content store -> ready.
func (s *server) processUpload(id string) {
	job := filepath.Join(s.jobs, id)
	uploadPath := filepath.Join(job, uploadFile)
	hlsPath := filepath.Join(job, "hls")
	if err := os.Mkdir(hlsPath, 0755); err != nil {
		slog.Error("creating hls dir", "id", id, "error", err)
		s.metadata.MarkFailed(id)
		return
	}
	defer os.RemoveAll(job)
	// 1. Transcode the upload into an HLS rendition under job/hls.
	slog.Info("transcode started", "id", id)
	if err := transcode.ToHLS(uploadPath, hlsPath); err != nil {
		slog.Error("transcoding failed", "id", id, "error", err)
		s.metadata.MarkFailed(id)
		return
	}
	slog.Info("transcode finished", "id", id)
	// 2. Publish the hls/ files into the content store under id.
	entries, err := os.ReadDir(hlsPath)
	if err != nil {
		slog.Error("reading hls dir", "id", id, "error", err)
		s.metadata.MarkFailed(id)
		return
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(hlsPath, entry.Name()))
		if err != nil {
			slog.Error("reading hls file", "id", id, "file", entry.Name(), "error", err)
			s.metadata.MarkFailed(id)
			return
		}
		if err := s.content.Write(id, entry.Name(), data); err != nil {
			slog.Error("publishing to content store", "id", id, "file", entry.Name(), "error", err)
			s.metadata.MarkFailed(id)
			return
		}
	}
	// 3. Extract a thumbnail and publish it (best-effort — a miss shouldn't fail the job).
	thumbPath := filepath.Join(job, "thumb.jpg")
	if err := transcode.Thumbnail(uploadPath, thumbPath); err != nil {
		slog.Warn("thumbnail extraction failed", "id", id, "error", err)
	} else if data, err := os.ReadFile(thumbPath); err != nil {
		slog.Warn("reading thumbnail", "id", id, "error", err)
	} else if err := s.content.Write(id, "thumb.jpg", data); err != nil {
		slog.Warn("publishing thumbnail", "id", id, "error", err)
	}
	// 4. Record duration (fall back to 0) and mark the video ready.
	duration, err := transcode.Duration(uploadPath)
	if err != nil {
		slog.Warn("duration probe failed", "id", id, "error", err)
	}
	s.metadata.MarkReady(id, duration)
}

// Done
// GET /watch/{id} — render "watch.html" with the video's metadata. 404 if it doesn't exist.
func (s *server) handleWatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	slog.Info("streaming", "id", id)
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
