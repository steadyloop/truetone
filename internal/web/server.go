// internal/web/server.go
package web

import (
	"net"
	"net/http"
)

type server struct {
	metadataService VideoMetadataService
	contentService  VideoContentService
	mux             *http.ServeMux
}

// video is the shape indexHTML expects per item ({{.EscapedId}}, {{.Id}}, {{.UploadTime}}).
type video struct {
	Id         string
	EscapedId  string
	UploadTime string
}

func NewServer(metadataService VideoMetadataService, contentService VideoContentService) *server {
	return &server{metadataService: metadataService, contentService: contentService}
}

func (s *server) Start(lis net.Listener) error {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/videos/", s.handleVideo)
	s.mux.HandleFunc("/content/", s.handleVideoContent)
	return http.Serve(lis, s.mux)
}

// GET / — render indexHTML (text/template over the const in templates.go) with []video
// from metadataService.List(). EscapedId = url.PathEscape(id). 200 on success, 500 on error.
func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// TODO
	http.Error(w, "not implemented", http.StatusNotImplemented)
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
