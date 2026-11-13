package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"prj_withdrawal/internal/bank"
)

type server struct {
	service *bank.Service
}

type transactionRequest struct {
	Amount string `json:"amount"`
}

type apiResponse struct {
	State       *bank.State       `json:"state,omitempty"`
	Transaction *bank.Transaction `json:"transaction,omitempty"`
	Error       string            `json:"error,omitempty"`
}

func NewServer(service *bank.Service) http.Handler {
	s := &server{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/transactions", s.transactions)
	mux.HandleFunc("/api/withdraw", s.transact(bank.TransactionWithdrawal))
	mux.HandleFunc("/api/deposit", s.transact(bank.TransactionDeposit))
	return mux
}

func (s *server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, page); err != nil {
		log.Printf("escribir página principal: %v", err)
	}
}

func (s *server) state(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "método no permitido"})
		return
	}
	state, err := s.service.Snapshot()
	if err != nil {
		log.Printf("consultar estado: %v", err)
		writeJSON(w, http.StatusInternalServerError, apiResponse{Error: "no se pudo leer el estado de la cuenta"})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{State: &state})
}

func (s *server) transactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "método no permitido"})
		return
	}
	state, err := s.service.Snapshot()
	if err != nil {
		log.Printf("consultar transacciones: %v", err)
		writeJSON(w, http.StatusInternalServerError, apiResponse{Error: "no se pudo leer el historial"})
		return
	}
	writeJSON(w, http.StatusOK, map[string][]bank.Transaction{"transactions": state.Transactions})
}

func (s *server) transact(kind bank.TransactionType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "método no permitido"})
			return
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			tx, err := s.recordInvalidRequest(kind)
			if err != nil {
				log.Printf("registrar contenido no admitido: %v", err)
				writeJSON(w, http.StatusInternalServerError, apiResponse{Error: "no se pudo registrar el error de la solicitud"})
				return
			}
			writeJSON(w, http.StatusUnsupportedMediaType, apiResponse{Transaction: &tx, Error: "el contenido debe ser JSON"})
			return
		}

		var request transactionRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			tx, recordErr := s.recordInvalidRequest(kind)
			if recordErr != nil {
				log.Printf("registrar JSON inválido: %v", recordErr)
				writeJSON(w, http.StatusInternalServerError, apiResponse{Error: "no se pudo registrar el error de la solicitud"})
				return
			}
			writeJSON(w, http.StatusBadRequest, apiResponse{Transaction: &tx, Error: "solicitud JSON inválida"})
			return
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			tx, recordErr := s.recordInvalidRequest(kind)
			if recordErr != nil {
				log.Printf("registrar JSON adicional: %v", recordErr)
				writeJSON(w, http.StatusInternalServerError, apiResponse{Error: "no se pudo registrar el error de la solicitud"})
				return
			}
			writeJSON(w, http.StatusBadRequest, apiResponse{Transaction: &tx, Error: "la solicitud debe contener un solo objeto JSON"})
			return
		}

		amount, parseErr := bank.ParseAmount(request.Amount)
		if parseErr != nil {
			amount = 0
		}
		var tx bank.Transaction
		var err error
		if kind == bank.TransactionWithdrawal {
			tx, err = s.service.Withdraw(bank.DefaultAccountID, amount)
		} else {
			tx, err = s.service.Deposit(bank.DefaultAccountID, amount)
		}

		if err != nil {
			if errors.Is(err, bank.ErrInvalidAmount) && parseErr != nil {
				writeJSON(w, http.StatusBadRequest, apiResponse{Transaction: &tx, Error: fmt.Sprintf("importe inválido: %v", parseErr)})
				return
			}
			if errors.Is(err, bank.ErrInvalidAmount) || errors.Is(err, bank.ErrAccountNotFound) || errors.Is(err, bank.ErrBalanceOverflow) {
				writeJSON(w, http.StatusBadRequest, apiResponse{Transaction: &tx, Error: err.Error()})
				return
			}
			log.Printf("procesar transacción: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiResponse{Transaction: &tx, Error: "no se pudo guardar la transacción"})
			return
		}
		if tx.Status == bank.StatusDeclined {
			writeJSON(w, http.StatusConflict, apiResponse{Transaction: &tx, Error: tx.Message})
			return
		}
		writeJSON(w, http.StatusOK, apiResponse{Transaction: &tx})
	}
}

func (s *server) recordInvalidRequest(kind bank.TransactionType) (bank.Transaction, error) {
	var tx bank.Transaction
	var err error
	if kind == bank.TransactionWithdrawal {
		tx, err = s.service.Withdraw(bank.DefaultAccountID, 0)
	} else {
		tx, err = s.service.Deposit(bank.DefaultAccountID, 0)
	}
	if errors.Is(err, bank.ErrInvalidAmount) {
		return tx, nil
	}
	return tx, err
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("escribir respuesta JSON: %v", err)
	}
}

const page = `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>PRJ Withdrawal - Operaciones bancarias</title>
  <style>
    :root { color-scheme: light; font-family: system-ui, sans-serif; background: #f3f6fb; color: #172033; }
    body { margin: 0; }
    main { max-width: 760px; margin: 3rem auto; padding: 0 1rem; }
    h1 { margin-bottom: .35rem; }
    .muted { color: #566176; }
    .card { background: white; border: 1px solid #dce3ee; border-radius: 12px; padding: 1.25rem; margin: 1rem 0; box-shadow: 0 4px 16px #1720330b; }
    .balance { font-size: 2rem; font-weight: 700; }
    .actions { display: grid; grid-template-columns: repeat(auto-fit, minmax(230px, 1fr)); gap: 1rem; }
    label { display: block; font-weight: 600; margin-bottom: .5rem; }
    input, button { box-sizing: border-box; width: 100%; padding: .7rem; border-radius: 8px; font: inherit; }
    input { border: 1px solid #aeb9ca; margin-bottom: .7rem; }
    button { border: 0; background: #155eef; color: white; font-weight: 700; cursor: pointer; }
    button:hover { background: #1049bd; }
    #message { min-height: 1.5rem; }
    .error { color: #b42318; }
    .success { color: #067647; }
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: .55rem; text-align: left; border-bottom: 1px solid #e6eaf0; }
    @media (max-width: 520px) { main { margin-top: 1.25rem; } th, td { font-size: .85rem; } }
  </style>
</head>
<body>
  <main>
    <h1>PRJ Withdrawal</h1>
    <p class="muted">Demo local para retiros y abonos con registro auditable.</p>
    <section class="card" aria-live="polite">
      <div id="owner" class="muted">Cuenta</div>
      <div id="balance" class="balance">Cargando...</div>
    </section>
    <p id="message" role="status" aria-live="polite"></p>
    <div class="actions">
      <form class="card" data-operation="withdraw">
        <h2>Retirar efectivo</h2>
        <label for="withdraw-amount">Importe</label>
        <input id="withdraw-amount" name="amount" inputmode="decimal" placeholder="0.00" required>
        <button type="submit">Retirar</button>
      </form>
      <form class="card" data-operation="deposit">
        <h2>Abonar efectivo</h2>
        <label for="deposit-amount">Importe</label>
        <input id="deposit-amount" name="amount" inputmode="decimal" placeholder="0.00" required>
        <button type="submit">Abonar</button>
      </form>
    </div>
    <section class="card">
      <h2>Transacciones</h2>
      <div id="history"></div>
    </section>
  </main>
  <script>
    const format = cents => (cents / 100).toLocaleString('es-PE', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    const message = document.querySelector('#message');
    async function refresh() {
      const response = await fetch('/api/state');
      if (!response.ok) throw new Error('No se pudo cargar la cuenta');
      const { state } = await response.json();
      document.querySelector('#owner').textContent = state.account.owner;
      document.querySelector('#balance').textContent = 'S/ ' + format(state.account.balance_cents);
      renderHistory(state.transactions);
    }
    function renderHistory(transactions) {
      const root = document.querySelector('#history');
      root.replaceChildren();
      if (transactions.length === 0) { root.textContent = 'Aún no hay operaciones.'; return; }
      const table = document.createElement('table');
      const head = document.createElement('tr');
      for (const label of ['Fecha UTC', 'Tipo', 'Resultado', 'Importe', 'Saldo']) {
        const cell = document.createElement('th'); cell.textContent = label; head.append(cell);
      }
      const thead = document.createElement('thead'); thead.append(head); table.append(thead);
      const body = document.createElement('tbody');
      for (const tx of [...transactions].reverse()) {
        const row = document.createElement('tr');
        for (const value of [new Date(tx.timestamp).toLocaleString(), tx.type, tx.status, 'S/ ' + format(tx.amount_cents), 'S/ ' + format(tx.balance_after_cents)]) {
          const cell = document.createElement('td'); cell.textContent = value; row.append(cell);
        }
        body.append(row);
      }
      table.append(body); root.append(table);
    }
    for (const form of document.querySelectorAll('form[data-operation]')) {
      form.addEventListener('submit', async event => {
        event.preventDefault();
        message.className = '';
        const amount = new FormData(form).get('amount');
        try {
          const response = await fetch('/api/' + form.dataset.operation, {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ amount })
          });
          const result = await response.json();
          message.textContent = result.error || result.transaction.message;
          message.className = response.ok ? 'success' : 'error';
          if (response.ok || response.status === 409 || result.transaction) await refresh();
        } catch (error) {
          message.textContent = error.message;
          message.className = 'error';
        }
      });
    }
    refresh().catch(error => { message.textContent = error.message; message.className = 'error'; });
  </script>
</body>
</html>`
