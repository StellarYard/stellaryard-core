package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/storage"
	"github.com/go-chi/chi/v5"
)

// Handlers holds dependencies for API handlers.
type Handlers struct {
	docker *docker.Client
	db     *storage.DB
}

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: struct {
			Code    string      `json:"code"`
			Message string      `json:"message"`
			Details interface{} `json:"details,omitempty"`
		}{Code: code, Message: message},
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// --- Container Handlers ---

func (h *Handlers) ListContainers(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.docker.ListStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, statuses)
}

func (h *Handlers) StartContainer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.docker.Start(r.Context(), name); err != nil {
		writeError(w, http.StatusBadRequest, "CONTAINER_START_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started", "name": name})
}

func (h *Handlers) StopContainer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.docker.Stop(r.Context(), name); err != nil {
		writeError(w, http.StatusBadRequest, "CONTAINER_STOP_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped", "name": name})
}

// --- Account Handlers ---

func (h *Handlers) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	acc := &storage.Account{
		ID:        generateID(),
		PublicKey: "G" + generateID()[:55], // Placeholder
		SecretKey: "S" + generateID()[:55], // Placeholder — testnet only
		Label:     req.Label,
		Network:   "local",
		CreatedAt: time.Now(),
	}

	if err := h.db.CreateAccount(acc); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, acc)
}

func (h *Handlers) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.db.ListAccounts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if accounts == nil {
		accounts = []storage.Account{}
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *Handlers) GetAccount(w http.ResponseWriter, r *http.Request) {
	publicKey := chi.URLParam(r, "publicKey")
	acc, err := h.db.GetAccount(publicKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if acc == nil {
		writeError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

// --- Contract Handlers ---

func (h *Handlers) DeployContract(w http.ResponseWriter, r *http.Request) {
	// V1: Accept WASM file upload, store deployment record
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "Contract deployment not yet implemented",
	})
}

func (h *Handlers) InvokeContract(w http.ResponseWriter, r *http.Request) {
	// V1: Accept contract ID and method, forward to Soroban RPC
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "Contract invocation not yet implemented",
	})
}

// --- Ledger Handlers ---

func (h *Handlers) GetLedgerSnapshot(w http.ResponseWriter, r *http.Request) {
	// V1: Query Horizon for current ledger state
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sequence":  0,
		"timestamp": time.Now(),
		"tx_count":  0,
		"note":      "Placeholder — will query Horizon in Phase 3",
	})
}

func (h *Handlers) ListTransactions(w http.ResponseWriter, r *http.Request) {
	// V1: Query Horizon for recent transactions
	writeJSON(w, http.StatusOK, []interface{}{})
}

// generateID generates a simple unique ID.
// In production, use UUID or similar.
func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}
