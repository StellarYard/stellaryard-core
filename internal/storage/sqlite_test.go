package storage

import (
	"path/filepath"
	"testing"
	"time"
)

// openTemp opens a migrated SQLite database in a throwaway directory.
func openTemp(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenRunsMigrations(t *testing.T) {
	db := openTemp(t)

	for _, table := range []string{"accounts", "contract_deployments"} {
		var name string
		err := db.conn.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %q not created by migrations: %v", table, err)
		}
		if name != table {
			t.Fatalf("expected table %q, got %q", table, name)
		}
	}
}

func TestOpenRunsMigrationsIdempotently(t *testing.T) {
	path := filepath.Join(t.TempDir(), "twice.db")

	first, err := Open(path)
	if err != nil {
		t.Fatalf("first Open() failed: %v", err)
	}
	first.Close()

	second, err := Open(path)
	if err != nil {
		t.Fatalf("second Open() on the same file failed: %v", err)
	}
	second.Close()
}

func newAccount(id, pub, label string, created time.Time) *Account {
	return &Account{
		ID:        id,
		PublicKey: pub,
		Label:     label,
		Network:   "local",
		CreatedAt: created,
	}
}

func TestCreateAndGetAccountRoundTrip(t *testing.T) {
	db := openTemp(t)
	created := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	in := newAccount("id-1", "GPUB1", "deployer", created)
	if err := db.CreateAccount(in); err != nil {
		t.Fatalf("CreateAccount() failed: %v", err)
	}

	got, err := db.GetAccount("GPUB1")
	if err != nil {
		t.Fatalf("GetAccount() failed: %v", err)
	}
	if got == nil {
		t.Fatal("GetAccount() returned nil for an account that exists")
	}

	if got.ID != in.ID || got.PublicKey != in.PublicKey || got.Label != in.Label ||
		got.Network != in.Network {
		t.Errorf("round trip mismatch:\n got  %+v\n want %+v", *got, *in)
	}
	if !got.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, created)
	}
}

func TestGetAccountMissingReturnsNilWithoutError(t *testing.T) {
	db := openTemp(t)

	got, err := db.GetAccount("GDOESNOTEXIST")
	if err != nil {
		t.Fatalf("GetAccount() on a missing key returned an error: %v", err)
	}
	if got != nil {
		t.Fatalf("GetAccount() on a missing key = %+v, want nil", *got)
	}
}

func TestCreateAccountRejectsDuplicateID(t *testing.T) {
	db := openTemp(t)
	now := time.Now()

	if err := db.CreateAccount(newAccount("dup", "GPUB-A", "a", now)); err != nil {
		t.Fatalf("first CreateAccount() failed: %v", err)
	}
	if err := db.CreateAccount(newAccount("dup", "GPUB-B", "b", now)); err == nil {
		t.Fatal("second CreateAccount() with the same primary key succeeded, want error")
	}
}

func TestCreateAccountRejectsDuplicatePublicKey(t *testing.T) {
	db := openTemp(t)
	now := time.Now()

	if err := db.CreateAccount(newAccount("id-a", "GPUB-SAME", "a", now)); err != nil {
		t.Fatalf("first CreateAccount() failed: %v", err)
	}
	if err := db.CreateAccount(newAccount("id-b", "GPUB-SAME", "b", now)); err == nil {
		t.Fatal("second CreateAccount() with the same public key succeeded, want error")
	}
}

func TestListAccountsEmptyReturnsNilNotError(t *testing.T) {
	db := openTemp(t)

	got, err := db.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() on an empty table failed: %v", err)
	}
	if got != nil {
		t.Fatalf("ListAccounts() on an empty table = %+v, want nil", got)
	}
}

func TestListAccountsOrdersNewestFirst(t *testing.T) {
	db := openTemp(t)
	older := time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)

	if err := db.CreateAccount(newAccount("old", "GPUB-OLD", "older", older)); err != nil {
		t.Fatalf("CreateAccount(older) failed: %v", err)
	}
	if err := db.CreateAccount(newAccount("new", "GPUB-NEW", "newer", newer)); err != nil {
		t.Fatalf("CreateAccount(newer) failed: %v", err)
	}

	got, err := db.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(ListAccounts()) = %d, want 2", len(got))
	}
	if got[0].PublicKey != "GPUB-NEW" || got[1].PublicKey != "GPUB-OLD" {
		t.Errorf("order = [%s, %s], want [GPUB-NEW, GPUB-OLD]", got[0].PublicKey, got[1].PublicKey)
	}
}

func TestDeploymentRoundTrip(t *testing.T) {
	db := openTemp(t)
	created := time.Date(2026, 10, 8, 15, 30, 0, 0, time.UTC)

	in := &ContractDeployment{
		ID:         "dep-1",
		WASMHash:   "abcdef",
		ContractID: "CCONTRACT",
		DeployedBy: "GPUB1",
		Network:    "local",
		CreatedAt:  created,
	}
	if err := db.CreateDeployment(in); err != nil {
		t.Fatalf("CreateDeployment() failed: %v", err)
	}

	got, err := db.ListDeployments()
	if err != nil {
		t.Fatalf("ListDeployments() failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(ListDeployments()) = %d, want 1", len(got))
	}
	if got[0] != *in {
		t.Errorf("deployment mismatch:\n got  %+v\n want %+v", got[0], *in)
	}
}

func TestListDeploymentsEmptyReturnsNilNotError(t *testing.T) {
	db := openTemp(t)

	got, err := db.ListDeployments()
	if err != nil {
		t.Fatalf("ListDeployments() on an empty table failed: %v", err)
	}
	if got != nil {
		t.Fatalf("ListDeployments() on an empty table = %+v, want nil", got)
	}
}

func TestOperationsAfterCloseReturnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "closed.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	if _, err := db.ListAccounts(); err == nil {
		t.Error("ListAccounts() after Close() succeeded, want error")
	}
}

// TestMigrationPurgesLegacySecretKeys verifies that an existing database
// containing legacy plaintext secret keys has those keys purged when opened,
// while preserving all other account metadata (ID, public key, label, network, timestamp).
func TestMigrationPurgesLegacySecretKeys(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")

	// 1. Manually create a legacy database schema with a plaintext secret key
	dbDirect, err := Open(dbPath)
	if err != nil {
		t.Fatalf("setup Open failed: %v", err)
	}

	createdTime := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)
	// Directly insert a row simulating pre-P0 schema containing a plaintext secret key
	_, err = dbDirect.conn.Exec(
		`INSERT INTO accounts (id, public_key, secret_key, label, network, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"legacy-acc-1", "GBM4V7G4EXAMPLEPUBKEY", "SBSECRETTESTNETKEYPLAINTEXT", "my-legacy-account", "local", createdTime,
	)
	if err != nil {
		t.Fatalf("failed to insert legacy row: %v", err)
	}

	// Verify the legacy secret is actually present on disk before migration purge
	var initialSecret string
	err = dbDirect.conn.QueryRow(`SELECT secret_key FROM accounts WHERE id = ?`, "legacy-acc-1").Scan(&initialSecret)
	if err != nil || initialSecret != "SBSECRETTESTNETKEYPLAINTEXT" {
		t.Fatalf("initial secret not found: %v, got %q", err, initialSecret)
	}
	dbDirect.Close()

	// 2. Re-open the database with Open() which runs migrations
	reopenedDB, err := Open(dbPath)
	if err != nil {
		t.Fatalf("re-open failed: %v", err)
	}
	defer reopenedDB.Close()

	// 3. Verify the secret column is now empty in SQLite
	var purgedSecret string
	err = reopenedDB.conn.QueryRow(`SELECT secret_key FROM accounts WHERE id = ?`, "legacy-acc-1").Scan(&purgedSecret)
	if err != nil {
		t.Fatalf("query secret_key failed: %v", err)
	}
	if purgedSecret != "" {
		t.Errorf("expected secret_key to be purged to empty string, got: %q", purgedSecret)
	}

	// 4. Verify all other fields are preserved
	acc, err := reopenedDB.GetAccount("GBM4V7G4EXAMPLEPUBKEY")
	if err != nil || acc == nil {
		t.Fatalf("GetAccount failed: %v", err)
	}
	if acc.ID != "legacy-acc-1" {
		t.Errorf("ID = %q, want legacy-acc-1", acc.ID)
	}
	if acc.PublicKey != "GBM4V7G4EXAMPLEPUBKEY" {
		t.Errorf("PublicKey = %q, want GBM4V7G4EXAMPLEPUBKEY", acc.PublicKey)
	}
	if acc.Label != "my-legacy-account" {
		t.Errorf("Label = %q, want my-legacy-account", acc.Label)
	}
	if acc.Network != "local" {
		t.Errorf("Network = %q, want local", acc.Network)
	}
	if !acc.CreatedAt.Equal(createdTime) {
		t.Errorf("CreatedAt = %v, want %v", acc.CreatedAt, createdTime)
	}

	// 5. Test idempotency: re-running migrations does not cause error or corrupt state
	thirdDB, err := Open(dbPath)
	if err != nil {
		t.Fatalf("third Open failed: %v", err)
	}
	thirdDB.Close()
}
