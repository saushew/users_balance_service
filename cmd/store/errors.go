package store

import "errors"

var (
	// ErrRecordNotFound ...
	ErrRecordNotFound = errors.New("record not found")
	// ErrUnknownTxType ...
	ErrUnknownTxType = errors.New("unknown transaction type")
	// ErrNotEnoughFunds ...
	ErrNotEnoughFunds = errors.New("user with user_id=%d has not enough funds")
)
