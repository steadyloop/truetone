// internal/web/templates.go
package web

import (
	"embed"
	"fmt"
	"hash/fnv"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"dur": func(seconds int) string {
		return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
	},
	"thumb": func(id string) string {
		h := fnv.New32a()
		h.Write([]byte(id))
		return fmt.Sprintf("t%d", (h.Sum32()%5)+1)
	},
}).ParseFS(templateFS, "templates/*.html"))
