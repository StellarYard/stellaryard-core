package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Account represents a managed test account.
//
// JSON tags follow api/openapi.yaml (camelCase). SecretKey is deliberately
// excluded from serialization: the OpenAPI spec does not promise it, no
// client consumes it, and returning key material over the API would be a
// needless exposure even for testnet credentials.
type Account struct {
	ID        string    `json:"id"`
	PublicKey string    `json:"publicKey"`
	SecretKey string    `json:"-"` // testnet only; never serialized
	Label     string    `json:"label"`
	Network   string    `json:"network"` // "local" | "testnet"
	CreatedAt time.Time `json:"createdAt"`
}

// ContractDeployment represents a deployed Soroban contract.
// JSON tags follow api/openapi.yaml (camelCase).
type ContractDeployment struct {
	ID         string    `json:"id"`
	WASMHash   string    `json:"wasmHash"`
	ContractID string    `json:"contractId"`
	DeployedBy string    `json:"deployedBy"`
	Network    string    `json:"network"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Open opens or creates a SQLite database and runs migrations.
func Open(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate runs database migrations.
func (db *DB) migrate() error {
	for _, m := range migrations() {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %s: %w", m[:50], err)
		}
	}
	return nil
}

// CreateAccount inserts a new account record.
func (db *DB) CreateAccount(acc *Account) error {
	_, err := db.conn.Exec(
		`INSERT INTO accounts (id, public_key, secret_key, label, network, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		acc.ID, acc.PublicKey, acc.SecretKey, acc.Label, acc.Network, acc.CreatedAt,
	)
	return err
}

// GetAccount returns an account by public key.
func (db *DB) GetAccount(publicKey string) (*Account, error) {
	acc := &Account{}
	err := db.conn.QueryRow(
		`SELECT id, public_key, secret_key, label, network, created_at
		 FROM accounts WHERE public_key = ?`, publicKey,
	).Scan(&acc.ID, &acc.PublicKey, &acc.SecretKey, &acc.Label, &acc.Network, &acc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return acc, err
}

// ListAccounts returns all managed accounts.
func (db *DB) ListAccounts() ([]Account, error) {
	rows, err := db.conn.Query(
		`SELECT id, public_key, secret_key, label, network, created_at FROM accounts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.ID, &acc.PublicKey, &acc.SecretKey, &acc.Label, &acc.Network, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, rows.Err()
}

// CreateDeployment inserts a new contract deployment record.
func (db *DB) CreateDeployment(dep *ContractDeployment) error {
	_, err := db.conn.Exec(
		`INSERT INTO contract_deployments (id, wasm_hash, contract_id, deployed_by, network, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		dep.ID, dep.WASMHash, dep.ContractID, dep.DeployedBy, dep.Network, dep.CreatedAt,
	)
	return err
}

// ListDeployments returns all contract deployments.
func (db *DB) ListDeployments() ([]ContractDeployment, error) {
	rows, err := db.conn.Query(
		`SELECT id, wasm_hash, contract_id, deployed_by, network, created_at
		 FROM contract_deployments ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []ContractDeployment
	for rows.Next() {
		var dep ContractDeployment
		if err := rows.Scan(&dep.ID, &dep.WASMHash, &dep.ContractID, &dep.DeployedBy, &dep.Network, &dep.CreatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}
