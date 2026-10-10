package handler

import (
	"context"
	"net/http"

	"github.com/meeptaquelle/backend/config"
	"github.com/meeptaquelle/backend/internal/server"
	"github.com/meeptaquelle/backend/storage"
)

var app http.Handler

func init() {
	cfg, err := config.Load()
	if err != nil {
		panic("config: " + err.Error())
	}

	store, err := storage.NewGoogleSheetsStoreVercel(
		cfg.SpreadsheetID, cfg.CredentialsB64, cfg.CredentialsPath,
	)
	if err != nil {
		panic("storage: " + err.Error())
	}

	// Vercel enforces its own timeout — don't add another one.
	app = server.Build(cfg, store, false)
	_ = context.Background
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
