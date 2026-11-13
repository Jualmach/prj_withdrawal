package cli

import (
	"bytes"
	"strings"
	"testing"

	"prj_withdrawal/internal/bank"
)

type testStore struct {
	state bank.State
}

func (s *testStore) Load() (bank.State, error) {
	return s.state, nil
}

func (s *testStore) Save(state bank.State) error {
	s.state = state
	return nil
}

func TestRunSupportsWithdrawalDepositBalanceAndHistory(t *testing.T) {
	store := &testStore{state: bank.State{
		Account: bank.Account{ID: bank.DefaultAccountID, Owner: "Ana", BalanceCents: 10000},
	}}
	service := bank.NewService(store)
	input := strings.NewReader("1\n35.50\n2\n25.75\n3\n4\n0\n")
	var output bytes.Buffer

	if err := Run(input, &output, service, bank.DefaultAccountID); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if store.state.Account.BalanceCents != 9025 {
		t.Fatalf("balance = %d, want 9025", store.state.Account.BalanceCents)
	}
	if len(store.state.Transactions) != 2 {
		t.Fatalf("transactions = %d, want 2", len(store.state.Transactions))
	}
	for _, expected := range []string{"Retiro aprobado", "Abono aprobado", "90.25", "success"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestRunReportsInvalidMenuChoice(t *testing.T) {
	store := &testStore{state: bank.DefaultState()}
	var output bytes.Buffer

	if err := Run(strings.NewReader("x\n"), &output, bank.NewService(store), bank.DefaultAccountID); err != nil {
		t.Fatalf("Run returned error at EOF: %v", err)
	}
	if !strings.Contains(output.String(), "Opción no válida") {
		t.Fatalf("output does not report the invalid choice: %s", output.String())
	}
}
