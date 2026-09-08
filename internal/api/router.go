package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/signer"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

// NewRouter creates the v1 API router.
func NewRouter(db *storage.DB, s signer.Signer, dc *docker.Client) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))

	r.Route("/api/v1", func(r chi.Router) {
		// Containers
		r.Post("/containers/{name}/start", handleContainerStart(dc))
		r.Post("/containers/{name}/stop", handleContainerStop(dc))
		r.Get("/containers", handleContainerList(dc))
		r.Get("/containers/{name}/logs", handleContainerLogs(dc))

		// Accounts
		r.Post("/accounts", handleAccountCreate(db, s))
		r.Get("/accounts", handleAccountList(db))
		r.Get("/accounts/{publicKey}", handleAccountGet(db))

		// Contracts
		r.Post("/contracts/deploy", handleContractDeploy(db))
		r.Post("/contracts/{contractId}/invoke", handleContractInvoke(db))

		// Ledger
		r.Get("/ledger/snapshot", handleLedgerSnapshot(db))
		r.Get("/ledger/transactions", handleLedgerTransactions(db))
	})

	return r
}
