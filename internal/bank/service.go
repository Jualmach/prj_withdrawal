package bank

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	ErrInvalidAmount   = errors.New("el importe debe ser mayor que cero")
	ErrAccountNotFound = errors.New("la cuenta solicitada no existe")
	ErrBalanceOverflow = errors.New("el saldo supera el límite permitido")
)

type Store interface {
	Load() (State, error)
	Save(State) error
}

type Service struct {
	store Store
	mu    sync.Mutex
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Withdraw(accountID string, amountCents int64) (Transaction, error) {
	return s.transact(accountID, amountCents, TransactionWithdrawal)
}

func (s *Service) Deposit(accountID string, amountCents int64) (Transaction, error) {
	return s.transact(accountID, amountCents, TransactionDeposit)
}

func (s *Service) Snapshot() (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.Load()
}

func (s *Service) transact(accountID string, amountCents int64, kind TransactionType) (Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.store.Load()
	if err != nil {
		return Transaction{}, fmt.Errorf("cargar datos de la cuenta: %w", err)
	}

	tx := Transaction{
		AccountID:   accountID,
		Type:        kind,
		AmountCents: amountCents,
		Timestamp:   time.Now().UTC(),
	}
	tx.ID, err = newTransactionID()
	if err != nil {
		return Transaction{}, fmt.Errorf("generar identificador de transacción: %w", err)
	}

	if state.Account.ID != accountID {
		return s.recordError(state, tx, ErrAccountNotFound)
	}
	if amountCents <= 0 {
		return s.recordError(state, tx, ErrInvalidAmount)
	}

	switch kind {
	case TransactionWithdrawal:
		if amountCents > state.Account.BalanceCents {
			tx.Status = StatusDeclined
			tx.Message = "saldo insuficiente"
		} else {
			state.Account.BalanceCents -= amountCents
			tx.Status = StatusSuccess
			tx.Message = "retiro aprobado"
		}
	case TransactionDeposit:
		if amountCents > math.MaxInt64-state.Account.BalanceCents {
			return s.recordError(state, tx, ErrBalanceOverflow)
		}
		state.Account.BalanceCents += amountCents
		tx.Status = StatusSuccess
		tx.Message = "abono aprobado"
	default:
		return s.recordError(state, tx, fmt.Errorf("tipo de transacción no válido: %q", kind))
	}

	tx.BalanceAfterCents = state.Account.BalanceCents
	state.Transactions = append(state.Transactions, tx)
	if err := s.store.Save(state); err != nil {
		tx.Status = StatusError
		tx.Message = "no se pudo guardar la transacción"
		return tx, fmt.Errorf("guardar transacción de forma atómica: %w", err)
	}
	return tx, nil
}

func (s *Service) recordError(state State, tx Transaction, cause error) (Transaction, error) {
	tx.Status = StatusError
	tx.BalanceAfterCents = state.Account.BalanceCents
	tx.Message = cause.Error()
	state.Transactions = append(state.Transactions, tx)
	if err := s.store.Save(state); err != nil {
		return tx, fmt.Errorf("%w; además, no se pudo registrar el error: %v", cause, err)
	}
	return tx, cause
}

func newTransactionID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(id[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}
