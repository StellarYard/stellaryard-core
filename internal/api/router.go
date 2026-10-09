package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/signer"
	"github.com/StellarYard/stellaryard-core/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// defaultAllowedOrigins lists permitted local development origins.
var defaultAllowedOrigins = map[string]bool{
	"http://localhost:3000": true, // Dashboard Vite dev server
	"http://127.0.0.1:3000": true,
	"http://localhost:8080": true, // Core local UI/API
	"http://127.0.0.1:8080": true,
}

// isOriginAllowed checks if the Origin header matches approved development origins.
func isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	// Check custom environment configuration
	if custom := os.Getenv("STELLARYARD_ALLOWED_ORIGINS"); custom != "" {
		for _, o := range strings.Split(custom, ",") {
			if strings.TrimSpace(o) == origin {
				return true
			}
		}
	}
	return defaultAllowedOrigins[origin]
}

// NewRouter creates the chi router with all v1 endpoints.
// If apiKey is non-empty, all routes under /api/v1 require Authorization: Bearer <apiKey>.
// /health remains public for operational liveness probes.
func NewRouter(dockerClient *docker.Client, db *storage.DB, s signer.Signer, apiKey string) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/health"))

	// Restricted CORS middleware (rejects wildcard, requires approved origins)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if isOriginAllowed(origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.Header().Set("Vary", "Origin")
				} else {
					// Origin not allowed
					if r.Method == "OPTIONS" {
						w.WriteHeader(http.StatusForbidden)
						return
					}
				}
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	h := &Handlers{
		docker: dockerClient,
		db:     db,
		signer: s,
	}

	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware when configured
		r.Use(RequireAuth(apiKey))

		// Containers
		r.Get("/containers", h.ListContainers)
		r.Post("/containers/{name}/start", h.StartContainer)
		r.Post("/containers/{name}/stop", h.StopContainer)
		r.Get("/containers/{name}/logs", h.StreamContainerLogs)

		// Accounts
		r.Post("/accounts", h.CreateAccount)
		r.Get("/accounts", h.ListAccounts)
		r.Get("/accounts/{publicKey}", h.GetAccount)

		// Contracts
		r.Post("/contracts/deploy", h.DeployContract)
		r.Post("/contracts/{contractId}/invoke", h.InvokeContract)

		// Ledger
		r.Get("/ledger/snapshot", h.GetLedgerSnapshot)
		r.Get("/ledger/transactions", h.ListTransactions)
	})

	return r
}
