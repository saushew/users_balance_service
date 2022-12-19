package sqlstore

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/saushew/users-balance-service/cmd/model"
	"github.com/saushew/users-balance-service/cmd/store"
)

// TransactionRepository ...
type TransactionRepository struct {
	store *Store
}

// Create ...
func (r *TransactionRepository) Create(t *model.Transaction) error {
	tx, err := r.store.db.Begin()
	if err != nil {
		return err
	}

	if err := insertTx(tx, t); err != nil {
		return err
	}

	if err := updateBalance(tx, t); err != nil {
		return err
	}
	return tx.Commit()
}

// Transfer ...
func (r *TransactionRepository) Transfer(t1, t2 *model.Transaction) error {
	tx, err := r.store.db.Begin()
	if err != nil {
		return err
	}

	if err := insertTx(tx, t1); err != nil {
		return err
	}

	if err := updateBalance(tx, t1); err != nil {
		return err
	}

	if err := insertTx(tx, t2); err != nil {
		return err
	}

	if err := updateBalance(tx, t2); err != nil {
		return err
	}
	return tx.Commit()
}

// Get ...
func (r *TransactionRepository) Get(userID, offset, limit int, order string) ([]*model.Transaction, error) {

	q := `SELECT * FROM transactions
	WHERE user_id = $1
	ORDER BY ts %s
	OFFSET $2 ROWS
	FETCH NEXT $3 ROWS ONLY;`

	if order == "desc" {
		q = fmt.Sprintf(q, "desc")
	} else {
		q = fmt.Sprintf(q, "asc")
	}

	rows, err := r.store.db.Query(
		q,
		userID,
		offset,
		limit,
	)
	if err != nil {
		return nil, err
	}

	resp := make([]*model.Transaction, 0, limit)
	for rows.Next() {
		t := &model.Transaction{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Type, &t.Details, &t.Timestamp); err != nil {
			return nil, err
		}
		resp = append(resp, t)
	}

	return resp, nil
}

func insertTx(tx *sql.Tx, t *model.Transaction) error {
	if err := tx.QueryRow(
		"INSERT INTO transactions (user_id, amount, tx_type, details, ts) VALUES ($1, $2, $3, $4, $5) RETURNING id;",
		t.UserID,
		t.Amount,
		t.Type,
		t.Details,
		t.Timestamp,
	).Scan(
		&t.ID,
	); err != nil {
		return RollbackTx(tx, err)
	}
	return nil
}

func updateBalance(tx *sql.Tx, t *model.Transaction) error {
	var balance float64

	if err := tx.QueryRow("SELECT balance FROM users WHERE id=$1", t.UserID).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err := tx.Exec("INSERT INTO users (id, balance) VALUES ($1, $2)", t.UserID, 0.0)
			if err != nil {
				return RollbackTx(tx, err)
			}
		} else {
			return RollbackTx(tx, err)
		}
	}

	if t.Type == model.TxDeposit {
		balance += t.Amount
	} else if t.Type == model.TxWithdraw {
		balance -= t.Amount
		if balance < 0 {
			err := fmt.Errorf(store.ErrNotEnoughFunds.Error(), t.UserID)
			return RollbackTx(tx, err)
		}
	} else {
		return RollbackTx(tx, store.ErrUnknownTxType)
	}

	if _, err := tx.Exec(
		"UPDATE users SET balance = $1 WHERE id = $2;",
		balance,
		t.UserID,
	); err != nil {
		return RollbackTx(tx, err)
	}
	return nil
}
