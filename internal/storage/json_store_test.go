package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"prj_withdrawal/internal/bank"
)

func TestJSONStoreCreatesDefaultStateWhenMissing(t *testing.T) {
	store := NewJSONStore(filepath.Join(t.TempDir(), "state.json"))

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if state.Account.ID != bank.DefaultAccountID || state.Account.BalanceCents != 100000 {
		t.Fatalf("unexpected default account: %+v", state.Account)
	}
}

func TestJSONStorePersistsBalanceAndTransactionsTogetherAsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	store := NewJSONStore(path)
	state := bank.DefaultState()
	state.Account.BalanceCents = 12575
	state.Transactions = []bank.Transaction{{
		ID:                "tx-1",
		AccountID:         state.Account.ID,
		Type:              bank.TransactionDeposit,
		Status:            bank.StatusSuccess,
		AmountCents:       2575,
		BalanceAfterCents: 12575,
	}}

	if err := store.Save(state); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), `"status": "success"`) {
		t.Fatalf("saved JSON does not contain the successful transaction: %s", data)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.Account.BalanceCents != 12575 || len(loaded.Transactions) != 1 {
		t.Fatalf("loaded state does not preserve balance and transaction: %+v", loaded)
	}
}

func TestJSONStoreDoesNotReplaceMalformedStateWithDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewJSONStore(path).Load(); err == nil {
		t.Fatal("Load must report malformed JSON instead of returning defaults")
	}
}
