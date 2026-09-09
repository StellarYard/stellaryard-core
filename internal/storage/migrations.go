package storage

// migrations returns the ordered list of SQL migration statements.
func migrations() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			public_key TEXT NOT NULL UNIQUE,
			secret_key TEXT NOT NULL,
			label TEXT NOT NULL DEFAULT '',
			network TEXT NOT NULL DEFAULT 'local',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS contract_deployments (
			id TEXT PRIMARY KEY,
			wasm_hash TEXT NOT NULL,
			contract_id TEXT NOT NULL,
			deployed_by TEXT NOT NULL,
			network TEXT NOT NULL DEFAULT 'local',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_network ON accounts(network)`,
		`CREATE INDEX IF NOT EXISTS idx_deployments_network ON contract_deployments(network)`,
	}
}
