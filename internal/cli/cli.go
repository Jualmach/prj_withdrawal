package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"prj_withdrawal/internal/bank"
)

func Run(input io.Reader, output io.Writer, service *bank.Service, accountID string) error {
	scanner := bufio.NewScanner(input)
	for {
		fmt.Fprintln(output, "\n=== PRJ Withdrawal ===")
		fmt.Fprintln(output, "1. Retirar efectivo")
		fmt.Fprintln(output, "2. Abonar efectivo")
		fmt.Fprintln(output, "3. Consultar saldo")
		fmt.Fprintln(output, "4. Ver transacciones")
		fmt.Fprintln(output, "0. Salir")
		fmt.Fprint(output, "Seleccione una opción: ")
		if !scanner.Scan() {
			return scanner.Err()
		}

		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			if err := runTransaction(scanner, output, service, accountID, bank.TransactionWithdrawal); err != nil {
				fmt.Fprintf(output, "ERROR: %v\n", err)
			}
		case "2":
			if err := runTransaction(scanner, output, service, accountID, bank.TransactionDeposit); err != nil {
				fmt.Fprintf(output, "ERROR: %v\n", err)
			}
		case "3":
			state, err := service.Snapshot()
			if err != nil {
				fmt.Fprintf(output, "ERROR: no se pudo consultar el saldo: %v\n", err)
				continue
			}
			fmt.Fprintf(output, "Saldo de %s: %s\n", state.Account.Owner, bank.FormatAmount(state.Account.BalanceCents))
		case "4":
			state, err := service.Snapshot()
			if err != nil {
				fmt.Fprintf(output, "ERROR: no se pudo consultar el historial: %v\n", err)
				continue
			}
			printTransactions(output, state.Transactions)
		case "0":
			fmt.Fprintln(output, "Hasta luego.")
			return nil
		default:
			fmt.Fprintln(output, "Opción no válida. Intente nuevamente.")
		}
	}
}

func runTransaction(scanner *bufio.Scanner, output io.Writer, service *bank.Service, accountID string, kind bank.TransactionType) error {
	if kind == bank.TransactionWithdrawal {
		fmt.Fprint(output, "Importe a retirar: ")
	} else {
		fmt.Fprint(output, "Importe a abonar: ")
	}
	if !scanner.Scan() {
		return scanner.Err()
	}

	amount, parseErr := bank.ParseAmount(scanner.Text())
	if parseErr != nil {
		amount = 0
	}

	var tx bank.Transaction
	var err error
	if kind == bank.TransactionWithdrawal {
		tx, err = service.Withdraw(accountID, amount)
	} else {
		tx, err = service.Deposit(accountID, amount)
	}
	if err != nil {
		if errors.Is(err, bank.ErrInvalidAmount) && parseErr != nil {
			fmt.Fprintf(output, "Operación registrada como error: %v\n", parseErr)
			return nil
		}
		return err
	}

	if tx.Status == bank.StatusDeclined {
		fmt.Fprintf(output, "Retiro rechazado: %s. Saldo: %s\n", tx.Message, bank.FormatAmount(tx.BalanceAfterCents))
		return nil
	}
	if tx.Type == bank.TransactionWithdrawal {
		fmt.Fprintf(output, "Retiro aprobado. Saldo restante: %s\n", bank.FormatAmount(tx.BalanceAfterCents))
	} else {
		fmt.Fprintf(output, "Abono aprobado. Nuevo saldo: %s\n", bank.FormatAmount(tx.BalanceAfterCents))
	}
	return nil
}

func printTransactions(output io.Writer, transactions []bank.Transaction) {
	if len(transactions) == 0 {
		fmt.Fprintln(output, "No hay transacciones registradas.")
		return
	}
	fmt.Fprintln(output, "Historial de transacciones:")
	for _, tx := range transactions {
		fmt.Fprintf(output, "%s | %s | %s | %s | saldo=%s | %s\n",
			tx.Timestamp.Format("2006-01-02 15:04:05 UTC"),
			tx.Type,
			tx.Status,
			bank.FormatAmount(tx.AmountCents),
			bank.FormatAmount(tx.BalanceAfterCents),
			tx.Message,
		)
	}
}
