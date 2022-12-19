package sqlstore

import (
	"database/sql"
	"fmt"
)

// RollbackTx ...
func RollbackTx(tx *sql.Tx, err error) error {
	rbErr := tx.Rollback()
	if rbErr != nil {
		return fmt.Errorf("tx err: %v; rollback err: %v", err, rbErr)
	}
	return fmt.Errorf("tx err: %v", err)
}
