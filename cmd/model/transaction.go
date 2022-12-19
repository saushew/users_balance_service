package model

type txType string

var (
	// TxWithdraw ...
	TxWithdraw = txType("withdraw")
	// TxDeposit ...
	TxDeposit = txType("deposit")
)

// Transaction ...
type Transaction struct {
	ID        int     `json:"tx_id" example:"123"`
	UserID    int     `json:"user_id" example:"1"`
	Amount    float64 `json:"amount" example:"12.34"`
	Type      txType  `json:"type" example:"deposit/withdraw"`
	Details   string  `json:"details" example:"some details"`
	Timestamp int64   `json:"timestamp" example:"1643756522"`
}
