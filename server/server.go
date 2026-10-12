package server

import (
	"context"
	"net/http"
	"time"

	"github.com/meeptaquelle/backend/config"
	"github.com/meeptaquelle/backend/handlers"
	"github.com/meeptaquelle/backend/services"
	"github.com/meeptaquelle/backend/storage"
)

// Build wires routes + middleware. Set withTimeout=true for local dev;
// leave false on Vercel (it enforces its own request timeout).
func Build(cfg *config.Config, store storage.Store, withTimeout bool) http.Handler {
	authSvc := services.NewAuthService(cfg.AuthPassword, cfg.AuthSecret, cfg.AuthTTL)
	authH := handlers.NewAuthHandler(authSvc)
	txH := handlers.NewTransactionHandler(services.NewTransactionService(store))
	catH := handlers.NewCategoryHandler(store)
	setH := handlers.NewSettingsHandler(store)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/transactions", txH.List)
	protected.HandleFunc("GET /api/transactions/{id}", txH.Get)
	protected.HandleFunc("POST /api/transactions", txH.Create)
	protected.HandleFunc("PUT /api/transactions/{id}", txH.Update)
	protected.HandleFunc("DELETE /api/transactions/{id}", txH.Delete)
	protected.HandleFunc("POST /api/transactions/{id}/complete", txH.Complete)
	protected.HandleFunc("GET /api/categories", catH.List)
	protected.HandleFunc("POST /api/categories", catH.Create)
	protected.HandleFunc("PUT /api/categories/{id}", catH.Update)
	protected.HandleFunc("POST /api/categories/{id}/archive", catH.Archive)
	protected.HandleFunc("GET /api/settings", setH.Get)
	protected.HandleFunc("PUT /api/settings", setH.Put)

	root := http.NewServeMux()
	root.HandleFunc("GET /api/health", handlers.Health)
	root.HandleFunc("POST /api/auth/login", authH.Login)
	root.Handle("/api/", handlers.RequireAuth(authSvc, protected))
	root.Handle("/", http.NotFoundHandler())

	var h http.Handler = root
	if withTimeout {
		h = WithTimeout(h, cfg.RequestTimeout)
	}
	h = WithLogging(h)
	h = WithCORS(h, cfg.AllowedOrigins)
	return h
}

func WithCORS(next http.Handler, allowed []string) http.Handler {
	set := make(map[string]bool, len(allowed))
	for _, o := range allowed {
		set[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && set[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func WithTimeout(next http.Handler, d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
