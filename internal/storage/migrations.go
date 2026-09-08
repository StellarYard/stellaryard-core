package storage

// Schema defines the initial SQLite schema for stellaryard-core.
//
// NOTE: ContainerStatus is NOT persisted — it's live-queried from Docker.
// This is an open question (see ROADMAP.md Phase 0) and may change.
const Schema = `
CREATE TABLE IF NOT EXISTS accounts (
	id TEXT PRIMARY KEY,
	public_key TEXT NOT NULL UNIQUE,
	secret_key TEXT NOT NULL,
	label TEXT NOT NULL,
	network TEXT NOT NULL DEFAULT 'testnet',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS contract_deployments (
	id TEXT PRIMARY KEY,
	wasm_hash TEXT NOT NULL,
	contract_id TEXT NOT NULL,
	deployed_by TEXT NOT NULL,
	network TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (deployed_by) REFERENCES accounts(public_key)
);

CREATE TABLE IF NOT EXISTS ledger_snapshots (
	sequence INTEGER PRIMARY KEY,
	timestamp DATETIME NOT NULL,
	tx_count INTEGER NOT NULL
);
`
