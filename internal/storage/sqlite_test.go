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
		SecretKey: "S-" + id,
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
		got.SecretKey != in.SecretKey || got.Network != in.Network {
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
