package storage

import "database/sql"

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
}

// New opens a SQLite database and runs migrations.
func New(path string) (*DB, error) {
	// TODO: implement
	// 1. Open SQLite with WAL mode
	// 2. Run migrations
	// 3. Return DB wrapper
	panic("not implemented")
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}
