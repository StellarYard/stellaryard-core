package api

import (
	"net/http"

	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates the chi router with all v1 endpoints.
func NewRouter(dockerClient *docker.Client, db *storage.DB) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/health"))

	// CORS for localhost development
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
	}

	r.Route("/api/v1", func(r chi.Router) {
		// Containers
		r.Get("/containers", h.ListContainers)
		r.Post("/containers/{name}/start", h.StartContainer)
		r.Post("/containers/{name}/stop", h.StopContainer)

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
