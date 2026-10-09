package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestContractDeploymentSchemaMismatch verifies that the error response
// from POST /api/v1/contracts/deploy does not follow the standard ErrorResponse envelope.
func TestContractDeploymentSchemaMismatch(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodPost, "/api/v1/contracts/deploy", `{}`)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}

	// Verify that it currently returns {"error": "..."} instead of {"error": {"code": "...", "message": "..."}}
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errVal, ok := raw["error"]
	if !ok {
		t.Fatalf("expected error key in response, got %v", raw)
	}

	// Current defect: error is a string, violating standard ErrorResponse struct
	if _, isStr := errVal.(string); !isStr {
		t.Errorf("expected string error value highlighting defect, got %T: %v", errVal, errVal)
	}
}

// TestListDeploymentsEndpointMissing confirms that GET /api/v1/contracts/deployments
// called by dashboard listDeployments() returns 404 because the route is missing.
func TestListDeploymentsEndpointMissing(t *testing.T) {
	h := newTestRouter(t)
	rec := doRequest(t, h, http.MethodGet, "/api/v1/contracts/deployments", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unimplemented route /api/v1/contracts/deployments, got %d", rec.Code)
	}
}
