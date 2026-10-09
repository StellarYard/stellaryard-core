package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/signer"
	"github.com/StellarYard/stellaryard-core/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// Handlers holds dependencies for API handlers.
type Handlers struct {
	docker *docker.Client
	db     *storage.DB
	signer signer.Signer
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

// validContainerNames contains allowed container identifiers per openapi.yaml
var validContainerNames = map[string]bool{
	"horizon":     true,
	"soroban-rpc": true,
}

func isValidContainer(name string) bool {
	return validContainerNames[name]
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
	if !isValidContainer(name) {
		writeError(w, http.StatusBadRequest, "INVALID_CONTAINER", "Unknown container name: "+name)
		return
	}
	if err := h.docker.Start(r.Context(), name); err != nil {
		writeError(w, http.StatusBadRequest, "CONTAINER_START_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started", "name": name})
}

func (h *Handlers) StopContainer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if !isValidContainer(name) {
		writeError(w, http.StatusBadRequest, "INVALID_CONTAINER", "Unknown container name: "+name)
		return
	}
	if err := h.docker.Stop(r.Context(), name); err != nil {
		writeError(w, http.StatusBadRequest, "CONTAINER_STOP_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped", "name": name})
}

// wsUpgrader configures gorilla/websocket upgrade with strict origin check.
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Non-browser clients (like CLI) omit Origin header — permit them.
		if origin == "" {
			return true
		}
		// Browser clients must match approved development origins.
		return isOriginAllowed(origin)
	},
}

// StreamContainerLogs streams container logs over WebSocket.
func (h *Handlers) StreamContainerLogs(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if !isValidContainer(name) {
		writeError(w, http.StatusBadRequest, "INVALID_CONTAINER", "Unknown container name: "+name)
		return
	}

	follow := true
	if f := r.URL.Query().Get("follow"); f == "false" {
		follow = false
	}
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100"
	}

	// Upgrade connection to WebSocket
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already writes HTTP error response if it fails
		return
	}
	defer conn.Close()

	// Context cancellation coordination
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Handle incoming client messages / ping-pong / close frames in a reader goroutine
	go func() {
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				// Client disconnected or sent close frame
				return
			}
		}
	}()

	// Retrieve container logs from Docker
	logReader, err := h.docker.Logs(ctx, name, docker.LogOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
	})
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error retrieving logs: "+err.Error()))
		return
	}
	defer logReader.Close()

	// Mutex to protect concurrent writes to WebSocket connection
	var writeMu sync.Mutex
	writeLine := func(line string) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteMessage(websocket.TextMessage, []byte(line))
	}

	// Stream demuxed lines to client
	if err := docker.StreamDemuxLines(ctx, logReader, writeLine); err != nil && err != context.Canceled {
		writeMu.Lock()
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.WriteMessage(websocket.TextMessage, []byte("Stream error: "+err.Error()))
		writeMu.Unlock()
	}

	// Clean close frame
	writeMu.Lock()
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "stream finished"))
	writeMu.Unlock()
}

// --- Account Handlers ---

func (h *Handlers) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Label   string `json:"label"`
		Network string `json:"network"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	network := req.Network
	if network == "" {
		network = "local"
	}

	// Generate keypair via the Signer boundary.
	// Raw secret keys NEVER enter this handler or SQLite.
	pubKey, err := h.signer.GenerateKeypair(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SIGNER_ERROR", "Failed to generate keypair: "+err.Error())
		return
	}

	acc := &storage.Account{
		ID:        generateID(),
		PublicKey: pubKey,
		Label:     req.Label,
		Network:   network,
		CreatedAt: time.Now().UTC(),
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

// generateID returns a random 32-character hexadecimal identifier.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand has no failure path on the platforms we target; fall
		// back to a timestamp so the request still completes rather than
		// panicking.
		return hex.EncodeToString([]byte(time.Now().UTC().Format("20060102150405.000000000")))[:32]
	}
	return hex.EncodeToString(b)
}
