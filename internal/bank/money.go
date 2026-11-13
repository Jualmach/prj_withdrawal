package bank

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseAmount(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("el importe es obligatorio")
	}

	negative := strings.HasPrefix(value, "-")
	unsigned := strings.TrimPrefix(value, "-")
	parts := strings.Split(unsigned, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("importe inválido: use el formato 12.34")
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) > 2 || fraction == "" {
			return 0, fmt.Errorf("importe inválido: use como máximo dos decimales")
		}
	}
	if len(fraction) == 1 {
		fraction += "0"
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("importe inválido: %w", err)
	}
	if whole > (1<<63-1)/100 {
		return 0, fmt.Errorf("importe fuera de rango")
	}

	cents := whole * 100
	if fraction != "" {
		fractionValue, parseErr := strconv.ParseInt(fraction, 10, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("importe inválido: %w", parseErr)
		}
		if cents > (1<<63-1)-fractionValue {
			return 0, fmt.Errorf("importe fuera de rango")
		}
		cents += fractionValue
	}
	if negative {
		cents = -cents
	}
	return cents, nil
}

func FormatAmount(cents int64) string {
	sign := ""
	magnitude := uint64(cents)
	if cents < 0 {
		sign = "-"
		magnitude = uint64(-(cents + 1)) + 1
	}
	return fmt.Sprintf("%s%d.%02d", sign, magnitude/100, magnitude%100)
}
