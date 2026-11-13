package bank

import "time"

const DefaultAccountID = "acct-001"

type Account struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	BalanceCents int64  `json:"balance_cents"`
}

type TransactionType string

const (
	TransactionWithdrawal TransactionType = "withdrawal"
	TransactionDeposit    TransactionType = "deposit"
)

type TransactionStatus string

const (
	StatusSuccess  TransactionStatus = "success"
	StatusDeclined TransactionStatus = "declined"
	StatusError    TransactionStatus = "error"
)

type Transaction struct {
	ID                string            `json:"id"`
	AccountID         string            `json:"account_id"`
	Type              TransactionType   `json:"type"`
	Status            TransactionStatus `json:"status"`
	AmountCents       int64             `json:"amount_cents"`
	BalanceAfterCents int64             `json:"balance_after_cents"`
	Timestamp         time.Time         `json:"timestamp"`
	Message           string            `json:"message,omitempty"`
}

type State struct {
	Account      Account       `json:"account"`
	Transactions []Transaction `json:"transactions"`
}

func DefaultState() State {
	return State{
		Account: Account{
			ID:           DefaultAccountID,
			Owner:        "Cliente",
			BalanceCents: 100000,
		},
		Transactions: []Transaction{},
	}
}
