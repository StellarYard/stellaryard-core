package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/models"
	"github.com/StellarYard/stellaryard-core/internal/signer"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

// --- Container handlers ---

func handleContainerStart(dc *docker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		// TODO: implement - call dc.Start(name)
		_ = name
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func handleContainerStop(dc *docker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		// TODO: implement - call dc.Stop(name)
		_ = name
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func handleContainerList(dc *docker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement - call dc.ListStatus()
		var statuses []models.ContainerStatus
		_ = statuses
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func handleContainerLogs(dc *docker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		// TODO: implement - upgrade to WebSocket, stream logs
		_ = name
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

// --- Account handlers ---

func handleAccountCreate(db *storage.DB, s signer.Signer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement
		// 1. Parse CreateAccountRequest
		// 2. Generate keypair via signer
		// 3. Fund account via Friendbot/local genesis
		// 4. Store in SQLite
		// 5. Return Account
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func handleAccountList(db *storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement - query all accounts from SQLite
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func handleAccountGet(db *storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicKey := chi.URLParam(r, "publicKey")
		// TODO: implement - query account by public key + fetch balance from Horizon
		_ = publicKey
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

// --- Contract handlers ---

func handleContractDeploy(db *storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement
		// 1. Parse WASM from request
		// 2. Deploy to local network via Soroban RPC
		// 3. Store deployment record in SQLite
		// 4. Return ContractDeployment
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func handleContractInvoke(db *storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contractId := chi.URLParam(r, "contractId")
		// TODO: implement
		// 1. Parse InvokeContractRequest
		// 2. Build XDR, sign via Signer interface
		// 3. Submit to Soroban RPC
		// 4. Return result
		_ = contractId
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

// --- Ledger handlers ---

func handleLedgerSnapshot(db *storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement - fetch current ledger state from Horizon
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func handleLedgerTransactions(db *storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement - paginated transaction list from Horizon
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}
