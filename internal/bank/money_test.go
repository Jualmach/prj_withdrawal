package bank

import "testing"

func TestParseAndFormatAmount(t *testing.T) {
	tests := []struct {
		input     string
		want      int64
		formatted string
	}{
		{input: "0", want: 0, formatted: "0.00"},
		{input: "12", want: 1200, formatted: "12.00"},
		{input: "12.3", want: 1230, formatted: "12.30"},
		{input: "12.34", want: 1234, formatted: "12.34"},
		{input: "-0.01", want: -1, formatted: "-0.01"},
	}
	for _, test := range tests {
		got, err := ParseAmount(test.input)
		if err != nil {
			t.Errorf("ParseAmount(%q) returned error: %v", test.input, err)
			continue
		}
		if got != test.want {
			t.Errorf("ParseAmount(%q) = %d, want %d", test.input, got, test.want)
		}
		if formatted := FormatAmount(got); formatted != test.formatted {
			t.Errorf("FormatAmount(%d) = %q, want %q", got, formatted, test.formatted)
		}
	}
	if formatted := FormatAmount(-1 << 63); formatted != "-92233720368547758.08" {
		t.Errorf("FormatAmount(MinInt64) = %q, want %q", formatted, "-92233720368547758.08")
	}
}

func TestParseAmountRejectsInvalidPrecisionAndOverflow(t *testing.T) {
	for _, input := range []string{"", "1.234", "1.", "abc", "92233720368547759.00"} {
		if _, err := ParseAmount(input); err == nil {
			t.Errorf("ParseAmount(%q) must return an error", input)
		}
	}
}
