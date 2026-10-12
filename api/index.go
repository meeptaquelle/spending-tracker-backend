package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/meeptaquelle/backend/config"
	"github.com/meeptaquelle/backend/server"
	"github.com/meeptaquelle/backend/storage"
)

var (
	once    sync.Once
	app     http.Handler
	initErr error
)

func setup() {
	cfg, err := config.Load()
	if err != nil {
		initErr = fmt.Errorf("config: %w", err)
		return
	}
	store, err := storage.NewGoogleSheetsStoreVercel(
		cfg.SpreadsheetID, cfg.CredentialsB64, cfg.CredentialsPath,
	)
	if err != nil {
		initErr = fmt.Errorf("storage: %w", err)
		return
	}
	// Vercel enforces its own timeout — don't add another one.
	app = server.Build(cfg, store, false)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Handler is the Vercel entrypoint. It never crashes on bad settings:
// it answers with a JSON error that says what is wrong.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(setup)

	if initErr != nil {
		if r.URL.Path == "/api/health" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"data": map[string]string{"status": "misconfigured", "problem": initErr.Error()},
			})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": map[string]string{"code": "MISCONFIGURED", "message": initErr.Error()},
		})
		return
	}

	app.ServeHTTP(w, r)
}
