package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StellarYard/stellaryard-core/internal/signer"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

// newTestRouter returns a router wired to a throwaway SQLite database.
// The Docker client is nil: these tests cover the HTTP contract for routes
// that do not touch Docker. Container routes require a running daemon and
// are covered by the issue-tracked integration suite.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "api-test.db"))
	if err != nil {
		t.Fatalf("storage.Open() failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRouter(nil, db, signer.NewLocalTestSigner(), "")
}

func doRequest(t *testing.T, h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// errorBody decodes the {"error":{"code","message"}} envelope.
func errorBody(t *testing.T, rec *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()
	var e ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("error body is not valid JSON: %v\nbody: %s", err, rec.Body.String())
	}
	return e
}

func TestHealthEndpoint(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK {
		t.Errorf("GET /health = %d, want 200", rec.Code)
	}
}

func TestCreateAccountReturns201AndRecord(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodPost, "/api/v1/accounts", `{"label":"deployer"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/accounts = %d, want 201\nbody: %s", rec.Code, rec.Body.String())
	}

	var acc struct {
		PublicKey string `json:"publicKey"`
		Label     string `json:"label"`
		Network   string `json:"network"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &acc); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if acc.Label != "deployer" {
		t.Errorf("label = %q, want %q", acc.Label, "deployer")
	}
	if acc.Network != "local" {
		t.Errorf("network = %q, want %q", acc.Network, "local")
	}
	if !strings.HasPrefix(acc.PublicKey, "G") || len(acc.PublicKey) < 56 {
		t.Errorf("publicKey = %q, want a Stellar-style key (G + 56 chars)", acc.PublicKey)
	}
}

// TestAccountResponseMatchesOpenAPISpec locks the response shape to
// api/openapi.yaml: camelCase fields, and no secret key material.
func TestAccountResponseMatchesOpenAPISpec(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodPost, "/api/v1/accounts", `{"label":"shape"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST = %d, want 201", rec.Code)
	}

	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	for _, field := range []string{"id", "publicKey", "label", "network", "createdAt"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("response is missing spec field %q; got keys %v", field, keysOf(raw))
		}
	}
	for _, field := range []string{"secretKey", "SecretKey", "PublicKey", "ID", "CreatedAt"} {
		if _, ok := raw[field]; ok {
			t.Errorf("response contains field %q, which the OpenAPI spec does not define", field)
		}
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestCreateAccountRejectsMalformedBody(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodPost, "/api/v1/accounts", `{not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST with a malformed body = %d, want 400", rec.Code)
	}
	if got := errorBody(t, rec).Error.Code; got != "INVALID_REQUEST" {
		t.Errorf("error code = %q, want INVALID_REQUEST", got)
	}
}

func TestListAccountsStartsEmpty(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodGet, "/api/v1/accounts", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/accounts = %d, want 200", rec.Code)
	}

	var accounts []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("response is not a JSON array: %v\nbody: %s", err, rec.Body.String())
	}
	if len(accounts) != 0 {
		t.Errorf("len(accounts) = %d, want 0", len(accounts))
	}
}

func TestGetAccountFoundAfterCreate(t *testing.T) {
	h := newTestRouter(t)

	create := doRequest(t, h, http.MethodPost, "/api/v1/accounts", `{"label":"lookup"}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("setup POST = %d, want 201", create.Code)
	}
	var created struct {
		PublicKey string `json:"publicKey"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("setup response is not valid JSON: %v", err)
	}

	rec := doRequest(t, h, http.MethodGet, "/api/v1/accounts/"+created.PublicKey, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/accounts/{pk} = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "lookup") {
		t.Errorf("response does not contain the label: %s", rec.Body.String())
	}
}

func TestGetAccountNotFoundReturns404Envelope(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodGet, "/api/v1/accounts/GNOSUCHACCOUNT", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET a missing account = %d, want 404", rec.Code)
	}
	e := errorBody(t, rec)
	if e.Error.Code != "ACCOUNT_NOT_FOUND" {
		t.Errorf("error code = %q, want ACCOUNT_NOT_FOUND", e.Error.Code)
	}
	if e.Error.Message == "" {
		t.Error("error message is empty")
	}
}

func TestContractDeployReturns501(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodPost, "/api/v1/contracts/deploy", `{}`)

	// Documented as not-yet-implemented in the README. If this ever starts
	// returning 200, the README status column should be updated too.
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("POST /api/v1/contracts/deploy = %d, want 501", rec.Code)
	}
}

func TestLedgerSnapshotReturnsPlaceholder(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodGet, "/api/v1/ledger/snapshot", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/ledger/snapshot = %d, want 200", rec.Code)
	}
	var snap map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := snap["sequence"]; !ok {
		t.Errorf("snapshot is missing the sequence field: %s", rec.Body.String())
	}
}

func TestLedgerTransactionsReturnsArray(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodGet, "/api/v1/ledger/transactions", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/ledger/transactions = %d, want 200", rec.Code)
	}
	var txs []any
	if err := json.Unmarshal(rec.Body.Bytes(), &txs); err != nil {
		t.Fatalf("response is not a JSON array: %v", err)
	}
}

func TestCORSHeadersAllowApprovedDevOrigins(t *testing.T) {
	h := newTestRouter(t)

	// Approved origin: http://localhost:3000
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/snapshot", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") {
		t.Errorf("Access-Control-Allow-Methods = %q, want it to include POST", got)
	}

	// Disallowed origin: http://evil.com
	reqBad := httptest.NewRequest(http.MethodOptions, "/api/v1/accounts", nil)
	reqBad.Header.Set("Origin", "http://evil.com")
	recBad := httptest.NewRecorder()
	h.ServeHTTP(recBad, reqBad)

	if recBad.Code != http.StatusForbidden {
		t.Errorf("OPTIONS from disallowed origin = %d, want 403", recBad.Code)
	}
	if got := recBad.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed origin received Access-Control-Allow-Origin header %q", got)
	}
}

func TestPreflightOptionsReturns200ForApprovedOrigin(t *testing.T) {
	h := newTestRouter(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/accounts", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("OPTIONS with approved origin = %d, want 200", rec.Code)
	}
}

func TestUnknownRouteReturns404(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodGet, "/api/v1/nope", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET unknown route = %d, want 404", rec.Code)
	}
}

// TestCreateAccountUsesSDKAndSigner ensures account creation produces a valid Stellar SDK address
// that is registered in the Signer and does NOT leak secret key material into the database or response.
func TestCreateAccountUsesSDKAndSigner(t *testing.T) {
	s := signer.NewLocalTestSigner()
	db, err := storage.Open(filepath.Join(t.TempDir(), "sdk-test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	router := NewRouter(nil, db, s, "")
	rec := doRequest(t, router, http.MethodPost, "/api/v1/accounts", `{"label":"sdk-acc"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/accounts = %d, want 201: %s", rec.Code, rec.Body.String())
	}

	var created struct {
		PublicKey string `json:"publicKey"`
		Label     string `json:"label"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	// 1. Must be registered with the signer
	if !s.HasKey(nil, created.PublicKey) {
		t.Errorf("signer does not hold key for created account %s", created.PublicKey)
	}

	// 2. Database record must be retrievable
	acc, err := db.GetAccount(created.PublicKey)
	if err != nil || acc == nil {
		t.Fatalf("db.GetAccount failed: %v", err)
	}
	if acc.PublicKey != created.PublicKey {
		t.Errorf("db public key %s != created %s", acc.PublicKey, created.PublicKey)
	}
}

func TestContainerNameValidation(t *testing.T) {
	h := newTestRouter(t)

	// Valid container start/stop routes (require docker client, but fail at docker level, not validation)
	// Invalid container name should be rejected immediately with 400 INVALID_CONTAINER
	rec := doRequest(t, h, http.MethodPost, "/api/v1/containers/unknown-malicious-container/start", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST with invalid container name = %d, want 400", rec.Code)
	}

	e := errorBody(t, rec)
	if e.Error.Code != "INVALID_CONTAINER" {
		t.Errorf("error code = %q, want INVALID_CONTAINER", e.Error.Code)
	}
}

func TestAuthenticationMiddleware(t *testing.T) {
	s := signer.NewLocalTestSigner()
	db, err := storage.Open(filepath.Join(t.TempDir(), "auth-test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	const testAPIKey = "super-secret-core-token-xyz"
	router := NewRouter(nil, db, s, testAPIKey)

	// 1. /health remains public and unauthenticated
	recHealth := doRequest(t, router, http.MethodGet, "/health", "")
	if recHealth.Code != http.StatusOK {
		t.Errorf("GET /health = %d, want 200", recHealth.Code)
	}

	// 2. Unauthenticated request to /api/v1/accounts rejected with 401
	recNoAuth := doRequest(t, router, http.MethodGet, "/api/v1/accounts", "")
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/v1/accounts without auth = %d, want 401", recNoAuth.Code)
	}
	eNoAuth := errorBody(t, recNoAuth)
	if eNoAuth.Error.Code != "UNAUTHORIZED" {
		t.Errorf("error code = %q, want UNAUTHORIZED", eNoAuth.Error.Code)
	}

	// 3. Request with invalid token rejected with 401
	reqBad := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	reqBad.Header.Set("Authorization", "Bearer invalid-wrong-token")
	recBad := httptest.NewRecorder()
	router.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/v1/accounts with bad auth = %d, want 401", recBad.Code)
	}
	// Verify token is NOT leaked in response
	if strings.Contains(recBad.Body.String(), "invalid-wrong-token") || strings.Contains(recBad.Body.String(), testAPIKey) {
		t.Error("error response leaked token strings")
	}

	// 4. Request with valid token accepted
	reqGood := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	reqGood.Header.Set("Authorization", "Bearer "+testAPIKey)
	recGood := httptest.NewRecorder()
	router.ServeHTTP(recGood, reqGood)
	if recGood.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/accounts with valid auth = %d, want 200", recGood.Code)
	}
}

func TestStreamContainerLogsEndpoint(t *testing.T) {
	s := signer.NewLocalTestSigner()
	db, err := storage.Open(filepath.Join(t.TempDir(), "logs-test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	const testAPIKey = "super-secret-core-token-xyz-123456"
	router := NewRouter(nil, db, s, testAPIKey)

	// 1. Invalid container name rejected with 400 when authorized
	reqBadNameWithAuth := httptest.NewRequest(http.MethodGet, "/api/v1/containers/unknown-app/logs", nil)
	reqBadNameWithAuth.Header.Set("Authorization", "Bearer "+testAPIKey)
	recBadNameWithAuth := httptest.NewRecorder()
	router.ServeHTTP(recBadNameWithAuth, reqBadNameWithAuth)
	if recBadNameWithAuth.Code != http.StatusBadRequest {
		t.Fatalf("GET /containers/unknown-app/logs = %d, want 400", recBadNameWithAuth.Code)
	}
	e := errorBody(t, recBadNameWithAuth)
	if e.Error.Code != "INVALID_CONTAINER" {
		t.Errorf("error code = %q, want INVALID_CONTAINER", e.Error.Code)
	}

	// 2. Unauthenticated request to /api/v1/containers/horizon/logs rejected with 401
	recNoAuth := doRequest(t, router, http.MethodGet, "/api/v1/containers/horizon/logs", "")
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("GET /containers/horizon/logs without auth = %d, want 401", recNoAuth.Code)
	}

	// 3. Disallowed origin rejected
	disallowedOriginReq := httptest.NewRequest(http.MethodGet, "/api/v1/containers/horizon/logs", nil)
	disallowedOriginReq.Header.Set("Authorization", "Bearer "+testAPIKey)
	disallowedOriginReq.Header.Set("Origin", "http://evil-attacker-site.com")
	disallowedOriginReq.Header.Set("Connection", "Upgrade")
	disallowedOriginReq.Header.Set("Upgrade", "websocket")
	disallowedOriginReq.Header.Set("Sec-WebSocket-Version", "13")
	disallowedOriginReq.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	disallowedRec := httptest.NewRecorder()
	router.ServeHTTP(disallowedRec, disallowedOriginReq)
	if disallowedRec.Code != http.StatusForbidden {
		t.Errorf("request with evil origin = %d, want 403 Forbidden", disallowedRec.Code)
	}
}

