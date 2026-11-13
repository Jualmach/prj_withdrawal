package bdd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

	"prj_withdrawal/internal/bank"
)

type scenarioStore struct {
	state bank.State
}

func (s *scenarioStore) Load() (bank.State, error) {
	return s.state, nil
}

func (s *scenarioStore) Save(state bank.State) error {
	s.state = state
	return nil
}

type scenario struct {
	service *bank.Service
	store   *scenarioStore
	tx      bank.Transaction
	err     error
}

func TestGherkinFeatures(t *testing.T) {
	features := filepath.Join("..", "..", "features")
	suite := godog.TestSuite{
		ScenarioInitializer: initializeScenario,
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{features},
			TestingT:  t,
			Strict:    true,
			Randomize: 0,
		},
	}
	if status := suite.Run(); status != 0 {
		t.Fatalf("Godog feature suite failed with status %d", status)
	}
}

func initializeScenario(ctx *godog.ScenarioContext) {
	state := &scenario{}
	ctx.Step(`^la cuenta de (.+) tiene un saldo de (\d+\.\d{2})$`, state.accountHasBalance)
	ctx.Step(`^(.+) retira (\d+\.\d{2})$`, state.ownerWithdraws)
	ctx.Step(`^(.+) abona (\d+\.\d{2}) en efectivo$`, state.ownerDeposits)
	ctx.Step(`^(.+) intenta abonar un importe de (\d+\.\d{2})$`, state.ownerDeposits)
	ctx.Step(`^el retiro es aprobado$`, state.withdrawalIsApproved)
	ctx.Step(`^el retiro es rechazado por saldo insuficiente$`, state.withdrawalIsDeclined)
	ctx.Step(`^el abono es aprobado$`, state.depositIsApproved)
	ctx.Step(`^el nuevo saldo es (\d+\.\d{2})$`, state.balanceIs)
	ctx.Step(`^el saldo permanece en (\d+\.\d{2})$`, state.balanceIs)
	ctx.Step(`^el resultado aprobado y el saldo (\d+\.\d{2}) quedan guardados juntos$`, state.successAndBalanceAreStored)
	ctx.Step(`^la transacción se registra en JSON$`, state.transactionRoundTripsAsJSON)
	ctx.Step(`^se registra el resultado rechazado en JSON$`, state.declineIsRecordedAsJSON)
	ctx.Step(`^se registra el resultado error en JSON$`, state.errorIsRecordedAsJSON)
}

func (s *scenario) accountHasBalance(owner, amount string) error {
	balance, err := bank.ParseAmount(amount)
	if err != nil {
		return err
	}
	s.store = &scenarioStore{state: bank.State{
		Account: bank.Account{ID: "scenario-account", Owner: owner, BalanceCents: balance},
	}}
	s.service = bank.NewService(s.store)
	return nil
}

func (s *scenario) ownerWithdraws(_ string, amount string) error {
	return s.runTransaction(amount, bank.TransactionWithdrawal)
}

func (s *scenario) ownerDeposits(_ string, amount string) error {
	return s.runTransaction(amount, bank.TransactionDeposit)
}

func (s *scenario) runTransaction(amount string, kind bank.TransactionType) error {
	cents, err := bank.ParseAmount(amount)
	if err != nil {
		return err
	}
	if kind == bank.TransactionWithdrawal {
		s.tx, s.err = s.service.Withdraw("scenario-account", cents)
	} else {
		s.tx, s.err = s.service.Deposit("scenario-account", cents)
	}
	return nil
}

func (s *scenario) withdrawalIsApproved() error {
	return s.assertStatus(bank.StatusSuccess)
}

func (s *scenario) withdrawalIsDeclined() error {
	return s.assertStatus(bank.StatusDeclined)
}

func (s *scenario) depositIsApproved() error {
	return s.assertStatus(bank.StatusSuccess)
}

func (s *scenario) assertStatus(want bank.TransactionStatus) error {
	if s.tx.Status != want {
		return fmt.Errorf("transaction status = %q, want %q (service error: %v)", s.tx.Status, want, s.err)
	}
	if want != bank.StatusDeclined && s.err != nil {
		return fmt.Errorf("transaction returned unexpected error: %w", s.err)
	}
	return nil
}

func (s *scenario) balanceIs(amount string) error {
	want, err := bank.ParseAmount(amount)
	if err != nil {
		return err
	}
	if got := s.store.state.Account.BalanceCents; got != want {
		return fmt.Errorf("balance = %s, want %s", bank.FormatAmount(got), bank.FormatAmount(want))
	}
	return nil
}

func (s *scenario) successAndBalanceAreStored(amount string) error {
	if err := s.assertStatus(bank.StatusSuccess); err != nil {
		return err
	}
	return s.balanceIs(amount)
}

func (s *scenario) transactionRoundTripsAsJSON() error {
	data, err := json.Marshal(s.store.state)
	if err != nil {
		return fmt.Errorf("encode transaction JSON: %w", err)
	}
	var decoded bank.State
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode transaction JSON: %w", err)
	}
	if decoded.Account.BalanceCents != s.store.state.Account.BalanceCents {
		return fmt.Errorf("JSON balance = %d, want %d", decoded.Account.BalanceCents, s.store.state.Account.BalanceCents)
	}
	if len(decoded.Transactions) != 1 || decoded.Transactions[0].Status != bank.StatusSuccess {
		return fmt.Errorf("JSON did not preserve the successful transaction: %+v", decoded.Transactions)
	}
	return nil
}

func (s *scenario) declineIsRecordedAsJSON() error {
	if err := s.assertStatus(bank.StatusDeclined); err != nil {
		return err
	}
	return s.assertJSONStatus(bank.StatusDeclined)
}

func (s *scenario) errorIsRecordedAsJSON() error {
	if s.tx.Status != bank.StatusError || s.err == nil {
		return fmt.Errorf("invalid amount should return and record an error; status=%q error=%v", s.tx.Status, s.err)
	}
	return s.assertJSONStatus(bank.StatusError)
}

func (s *scenario) assertJSONStatus(want bank.TransactionStatus) error {
	data, err := json.Marshal(s.store.state)
	if err != nil {
		return fmt.Errorf("encode transaction JSON: %w", err)
	}
	var decoded bank.State
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode transaction JSON: %w", err)
	}
	if len(decoded.Transactions) != 1 || decoded.Transactions[0].Status != want {
		return fmt.Errorf("JSON did not preserve transaction status %q: %+v", want, decoded.Transactions)
	}
	return nil
}
