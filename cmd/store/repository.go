package store

import (
	"github.com/saushew/users-balance-service/cmd/model"
)

// UserRepository ...
type UserRepository interface {
	Find(int) (*model.User, error)
}

// TransactionRepository ...
type TransactionRepository interface {
	Create(*model.Transaction) error
	Transfer(*model.Transaction, *model.Transaction) error
	Get(int, int, int, string) ([]*model.Transaction, error)
}
