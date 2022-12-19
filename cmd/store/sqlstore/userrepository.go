package sqlstore

import (
	"database/sql"

	"github.com/saushew/users-balance-service/cmd/model"
	"github.com/saushew/users-balance-service/cmd/store"
)

// UserRepository ...
type UserRepository struct {
	store *Store
}

// Find ...
func (r *UserRepository) Find(id int) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT balance FROM users WHERE id = $1",
		id,
	).Scan(
		&u.Balance,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	u.ID = id

	return u, nil
}
