// internal/web/server.go
// Defines the backend that uses the filesystem for content, sqlite fo metadata
package web

import (
	"html/template"
	"net"
	"net/http"
)

type server struct {
	Metadata MetadataStore
	Content  ContentStore
	mux      *http.ServeMux
}

func NewServer(metadata MetadataStore, content ContentStore) *server {
	return &server{Metadata: metadata, Content: content}
}

func (s *server) Start(lis net.Listener) error {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/videos/", s.handleVideo)
	s.mux.HandleFunc("/content/", s.handleVideoContent)
	return http.Serve(lis, s.mux)
}

// GET / - Renders the home screen (index), with []content from metadata.List()
func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	metadata, err := s.Metadata.List()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	if err := tmpl.Execute(w, metadata); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// POST /upload — multipart field "file". Video ID = filename minus ".mp4".
// 400 if the file field is missing · 409 if the ID already exists ·
// save to a temp dir -> ffmpeg -> DASH (command in the lab write-up) ->
// contentService.Write each output file -> metadataService.Create -> 303 redirect to /.
func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// TODO
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// GET /videos/{id} — render videoHTML with the video's metadata. 404 if it doesn't exist.
func (s *server) handleVideo(w http.ResponseWriter, r *http.Request) {
	// TODO: videoId := r.URL.Path[len("/videos/"):]
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// GET /content/{videoId}/{filename} — body = contentService.Read(...). 500 on any failure.
func (s *server) handleVideoContent(w http.ResponseWriter, r *http.Request) {
	// TODO: parse the two path parts, write the bytes
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
