package handler

import (
	"net/http"

	"github.com/meeptaquelle/backend/config"
	"github.com/meeptaquelle/backend/server"
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
}

// Handler is the Vercel entrypoint.
//
// vercel.json rewrites /api/<anything> to /api?__path=<anything>, because this
// single function lives at /api. Inside the function the URL path is therefore
// "/api", not the path the client asked for, so we rebuild it here.
func Handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if p, ok := q["__path"]; ok {
		q.Del("__path")
		r.URL.Path = "/api/" + p[0]
		r.URL.RawQuery = q.Encode()
	} else if r.URL.Path == "/api" {
		r.URL.Path = "/api/"
	}
	app.ServeHTTP(w, r)
}
