package sqlite

import "fmt"

// ValidateDB returns an error if the DB or its underlying connection is nil.
func ValidateDB(db *DB, caller string) error {
	if db == nil || db.Conn == nil {
		return fmt.Errorf("%s: nil db", caller)
	}
	return nil
}
