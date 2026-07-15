# Phase 0

**Goal:** single-node site that uploads an MP4, converts to DASH, and streams it back.
**Run:** `mkdir -p storage && go run ./cmd/web -port 8080 sqlite ./metadata.db fs ./storage`

---

### [x] 1. Metadata store — `internal/web/sqlite.go`  (SQLite)
- Updated schema to include title (in interfaces.go as well as sqlite.go)
- Implemented Create, List, Read

### [x] 2. Content store — `internal/web/fs.go`  (fs)
- implemented Write, Read - -07/14/2026

### [ ] 3. Wiring — `cmd/web/main.go`
- handled selection of metadata service (sqlite), content service (fs), and config file scheme. -07/15/2026
- [ ] checkpoint: `go build ./...` compiles and boots

### [ ] 4. HTTP handlers — `internal/web/server.go`  (simplest first)
- Implemented handleContent, handleIndex, handleWatch, handleUpload (07/14/2026)



---

## Tech Debt
1. Read up on errors on Go
2. Read up on sql on Go
    https://pkg.go.dev/database/sql#DB
3. Read up on structured logging (log/slog) — main.go, 07/15/2026
    - Idea: Why key-value instead of fmt.Println? Structured logs allow querying/scaling by making it grep/parseable.
    - https://pkg.go.dev/log/slog

---

## Done when
- [ ] upload an mp4 → redirects home, video shows in list
- [ ] click it → plays via dash.js
- [ ] re-upload same name → 409 · no file field → 400
- [ ] `metadata.db` + `storage/<id>/manifest.mpd` exist on disk
