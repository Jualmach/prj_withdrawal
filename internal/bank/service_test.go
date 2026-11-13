package bank

import (
	"errors"
	"sync"
	"testing"
)

type memoryStore struct {
	state State
	err   error
}

func (s *memoryStore) Load() (State, error) {
	return s.state, nil
}

func (s *memoryStore) Save(state State) error {
	if s.err != nil {
		return s.err
	}
	s.state = state
	return nil
}

func newTestService(t *testing.T, balance int64) (*Service, *memoryStore) {
	t.Helper()
	store := &memoryStore{state: State{
		Account: Account{ID: "acct-1", Owner: "Ana", BalanceCents: balance},
	}}
	return NewService(store), store
}

func TestUS01aWithdrawalApproved(t *testing.T) {
	service, store := newTestService(t, 10000)

	tx, err := service.Withdraw("acct-1", 3550)
	if err != nil {
		t.Fatalf("Withdraw returned error: %v", err)
	}
	if tx.Status != StatusSuccess {
		t.Fatalf("status = %q, want %q", tx.Status, StatusSuccess)
	}
	if store.state.Account.BalanceCents != 6450 {
		t.Fatalf("balance = %d, want 6450", store.state.Account.BalanceCents)
	}
	assertOneTransaction(t, store.state, TransactionWithdrawal, StatusSuccess, 3550, 6450)
}

func TestUS01aWithdrawalDeclinedWithoutChangingBalance(t *testing.T) {
	service, store := newTestService(t, 10000)

	tx, err := service.Withdraw("acct-1", 12000)
	if err != nil {
		t.Fatalf("Withdraw returned error: %v", err)
	}
	if tx.Status != StatusDeclined {
		t.Fatalf("status = %q, want %q", tx.Status, StatusDeclined)
	}
	if store.state.Account.BalanceCents != 10000 {
		t.Fatalf("balance = %d, want 10000", store.state.Account.BalanceCents)
	}
	assertOneTransaction(t, store.state, TransactionWithdrawal, StatusDeclined, 12000, 10000)
}

func TestUS01bCashDeposit(t *testing.T) {
	service, store := newTestService(t, 10000)

	tx, err := service.Deposit("acct-1", 2575)
	if err != nil {
		t.Fatalf("Deposit returned error: %v", err)
	}
	if tx.Status != StatusSuccess {
		t.Fatalf("status = %q, want %q", tx.Status, StatusSuccess)
	}
	if store.state.Account.BalanceCents != 12575 {
		t.Fatalf("balance = %d, want 12575", store.state.Account.BalanceCents)
	}
	assertOneTransaction(t, store.state, TransactionDeposit, StatusSuccess, 2575, 12575)
}

func TestUS04InvalidAmountIsLoggedAsErrorWithoutChangingBalance(t *testing.T) {
	service, store := newTestService(t, 10000)

	tx, err := service.Deposit("acct-1", 0)
	if err == nil {
		t.Fatal("Deposit with zero amount must return an error")
	}
	if tx.Status != StatusError {
		t.Fatalf("status = %q, want %q", tx.Status, StatusError)
	}
	if store.state.Account.BalanceCents != 10000 {
		t.Fatalf("balance = %d, want 10000", store.state.Account.BalanceCents)
	}
	assertOneTransaction(t, store.state, TransactionDeposit, StatusError, 0, 10000)
}

func TestUS04SaveFailureDoesNotReportTransactionAsSuccessful(t *testing.T) {
	service, store := newTestService(t, 10000)
	store.err = errors.New("disk full")

	tx, err := service.Deposit("acct-1", 500)
	if err == nil {
		t.Fatal("Deposit must return the persistence error")
	}
	if tx.Status != StatusError {
		t.Fatalf("status = %q, want %q", tx.Status, StatusError)
	}
	if got := store.state.Account.BalanceCents; got != 10000 {
		t.Fatalf("persisted balance = %d, want 10000", got)
	}
	if len(store.state.Transactions) != 0 {
		t.Fatalf("persisted transactions = %d, want 0", len(store.state.Transactions))
	}
}

func TestConcurrentDepositsKeepBalanceAndHistoryConsistent(t *testing.T) {
	service, store := newTestService(t, 10000)
	const operations = 32
	var waitGroup sync.WaitGroup
	errorsFound := make(chan error, operations)

	for range operations {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := service.Deposit("acct-1", 1)
			if err != nil {
				errorsFound <- err
			}
		}()
	}
	waitGroup.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Errorf("Deposit returned error: %v", err)
	}

	if store.state.Account.BalanceCents != 10000+operations {
		t.Fatalf("balance = %d, want %d", store.state.Account.BalanceCents, 10000+operations)
	}
	if len(store.state.Transactions) != operations {
		t.Fatalf("transactions = %d, want %d", len(store.state.Transactions), operations)
	}
	seen := make(map[string]bool, operations)
	for _, tx := range store.state.Transactions {
		if seen[tx.ID] {
			t.Errorf("duplicate transaction ID %q", tx.ID)
		}
		seen[tx.ID] = true
	}
}

func assertOneTransaction(t *testing.T, state State, kind TransactionType, status TransactionStatus, amount, balance int64) {
	t.Helper()
	if len(state.Transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(state.Transactions))
	}
	tx := state.Transactions[0]
	if tx.Type != kind || tx.Status != status || tx.AmountCents != amount || tx.BalanceAfterCents != balance {
		t.Fatalf("transaction = %+v, want type=%q status=%q amount=%d balance=%d", tx, kind, status, amount, balance)
	}
}
