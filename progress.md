# Phase 0

**Goal:** single-node site that uploads an MP4, converts to DASH, and streams it back.
**Run:** `mkdir -p storage && go run ./cmd/web -port 8080 sqlite ./metadata.db fs ./storage`

---

### [x] 1. Metadata store — `internal/web/sqlite.go`  (SQLite)
- Updated schema to include title (in interfaces.go as well as sqlite.go)
- Implemented Create, List, Read

## 2. Content store — `internal/web/fs.go`  (local filesystem)
- [ ] `Write` — `os.MkdirAll(baseDir/videoId)` then `os.WriteFile(.../filename)`
- [ ] `Read` — `os.ReadFile(baseDir/videoId/filename)`
- [ ] uncomment interface assertion

## 3. Wiring — `cmd/web/main.go`
- [ ] `case "sqlite":` → build metadata service (handle the error)
- [ ] `case "fs":` → build content service
- [ ] checkpoint: `go build ./...` compiles and boots

## 4. HTTP handlers — `internal/web/server.go`  (simplest first)
- [ ] `handleVideoContent`  GET `/content/{id}/{file}` — Read bytes → body; 500 on err
- [ ] `handleIndex`         GET `/` — `List()` → render `indexHTML`; `EscapedId = url.PathEscape(id)`
- [ ] `handleVideo`         GET `/videos/{id}` — `Read(id)`; 404 if nil; else render `videoHTML`
- [ ] `handleUpload`        POST `/upload` — see breakdown below

### handleUpload steps
- [ ] `r.FormFile("file")` — 400 if missing
- [ ] video id = filename minus `.mp4`
- [ ] `Read(id)` to reject dupes — 409 if exists
- [ ] save upload to temp dir (`os.MkdirTemp`)
- [ ] run ffmpeg → DASH (command in lab7.md)
- [ ] `contentService.Write` each output file (`manifest.mpd` + `.m4s`)
- [ ] `metadataService.Create` then 303 redirect to `/`

---

## Tech Debt
1. Read up on errors on Go
2. Read up on sql on Go
    https://pkg.go.dev/database/sql#DB

---

## Done when
- [ ] upload an mp4 → redirects home, video shows in list
- [ ] click it → plays via dash.js
- [ ] re-upload same name → 409 · no file field → 400
- [ ] `metadata.db` + `storage/<id>/manifest.mpd` exist on disk

## Deps / notes
- [ ] `ffmpeg` installed (`brew install ffmpeg`)
- go-sqlite3 is cgo (needs C compiler — you have clang)
