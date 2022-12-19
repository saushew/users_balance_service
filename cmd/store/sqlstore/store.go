package sqlstore

import (
	"database/sql"

	"github.com/saushew/users-balance-service/cmd/store"
)

// Store ...
type Store struct {
	db                    *sql.DB
	userRepository        *UserRepository
	transactionRepository *TransactionRepository
}

// New ...
func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// User ...
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}

// Transaction ...
func (s *Store) Transaction() store.TransactionRepository {
	if s.transactionRepository != nil {
		return s.transactionRepository
	}

	s.transactionRepository = &TransactionRepository{
		store: s,
	}

	return s.transactionRepository
}
