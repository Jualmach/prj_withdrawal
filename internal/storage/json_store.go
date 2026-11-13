package storage

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"prj_withdrawal/internal/bank"
)

//go:embed default-state.json
var defaultStateJSON []byte

type JSONStore struct {
	path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func (s *JSONStore) Load() (bank.State, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		data = defaultStateJSON
		err = nil
	}
	if err != nil {
		return bank.State{}, fmt.Errorf("leer %q: %w", s.path, err)
	}

	var state bank.State
	if err := json.Unmarshal(data, &state); err != nil {
		return bank.State{}, fmt.Errorf("decodificar %q: %w", s.path, err)
	}
	if state.Account.ID == "" || state.Account.BalanceCents < 0 {
		return bank.State{}, fmt.Errorf("el archivo %q contiene datos de cuenta inválidos", s.path)
	}
	if state.Transactions == nil {
		state.Transactions = []bank.Transaction{}
	}
	return state, nil
}

func (s *JSONStore) Save(state bank.State) error {
	if state.Account.ID == "" || state.Account.BalanceCents < 0 {
		return fmt.Errorf("se rechazó guardar una cuenta con datos inválidos")
	}
	if state.Transactions == nil {
		state.Transactions = []bank.Transaction{}
	}

	directory := filepath.Dir(s.path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("crear directorio de datos: %w", err)
	}
	temp, err := os.CreateTemp(directory, ".withdrawal-*.tmp")
	if err != nil {
		return fmt.Errorf("crear archivo temporal: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	encoder := json.NewEncoder(temp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		temp.Close()
		return fmt.Errorf("codificar estado JSON: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sincronizar archivo temporal: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("cerrar archivo temporal: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("reemplazar archivo de estado: %w", err)
	}
	return nil
}
