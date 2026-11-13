package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func newTestServer() (*httptest.Server, *testStore) {
	store := &testStore{state: bank.State{
		Account: bank.Account{ID: bank.DefaultAccountID, Owner: "Ana", BalanceCents: 10000},
	}}
	return httptest.NewServer(NewServer(bank.NewService(store))), store
}

func TestHomePageAndStateEndpoint(t *testing.T) {
	server, _ := newTestServer()
	defer server.Close()

	response, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET / returned error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	stateResponse, err := http.Get(server.URL + "/api/state")
	if err != nil {
		t.Fatalf("GET /api/state returned error: %v", err)
	}
	defer stateResponse.Body.Close()
	if stateResponse.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/state status = %d, want %d", stateResponse.StatusCode, http.StatusOK)
	}
}

func TestDepositEndpointPersistsSuccessfulTransaction(t *testing.T) {
	server, store := newTestServer()
	defer server.Close()

	response, err := http.Post(server.URL+"/api/deposit", "application/json", strings.NewReader(`{"amount":"25.75"}`))
	if err != nil {
		t.Fatalf("POST /api/deposit returned error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/deposit status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if store.state.Account.BalanceCents != 12575 || len(store.state.Transactions) != 1 {
		t.Fatalf("deposit was not persisted: %+v", store.state)
	}
}

func TestWithdrawalDeclineAndValidationErrorHaveDistinctResponses(t *testing.T) {
	server, store := newTestServer()
	defer server.Close()

	response, err := http.Post(server.URL+"/api/withdraw", "application/json", strings.NewReader(`{"amount":"120.00"}`))
	if err != nil {
		t.Fatalf("POST /api/withdraw returned error: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("declined withdrawal status = %d, want %d", response.StatusCode, http.StatusConflict)
	}

	response, err = http.Post(server.URL+"/api/deposit", "application/json", strings.NewReader(`{"amount":"0.00"}`))
	if err != nil {
		t.Fatalf("POST /api/deposit returned error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid deposit status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	if len(store.state.Transactions) != 2 || store.state.Transactions[1].Status != bank.StatusError {
		t.Fatalf("expected decline and error in log, got %+v", store.state.Transactions)
	}
}

func TestTransactionEndpointRejectsMalformedJSONAndLogsError(t *testing.T) {
	server, store := newTestServer()
	defer server.Close()

	response, err := http.Post(server.URL+"/api/deposit", "application/json", bytes.NewBufferString(`{"amount":`))
	if err != nil {
		t.Fatalf("POST /api/deposit returned error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed request status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if len(store.state.Transactions) != 1 || store.state.Transactions[0].Status != bank.StatusError {
		t.Fatalf("malformed request was not recorded as an error: %+v", store.state.Transactions)
	}
}
