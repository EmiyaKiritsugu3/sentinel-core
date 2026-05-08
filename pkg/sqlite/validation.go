package sqlite

import (
	"errors"
	"fmt"
)

// ErrNilDB is returned by ValidateDB when the DB or its connection is nil.
var ErrNilDB = errors.New("nil db")

// ValidateDB returns an error if the DB or its underlying connection is nil.
func ValidateDB(db *DB, caller string) error {
	if db == nil || db.Conn == nil {
		return fmt.Errorf("%s: %w", caller, ErrNilDB)
	}
	return nil
}
