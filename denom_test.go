package denom

import (
	"math"
	"testing"
)

func TestFromMajorString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		cur     Currency
		want    int64
		wantErr bool
	}{
		{"usd basic", "50023.44", USD, 5_002_344, false},
		{"usd positive sign", "+1.50", USD, 150, false},
		{"usd negative", "-0.50", USD, -50, false},
		{"scale zero", "123", KRW, 123, false},
		{"too many decimals", "1.234", USD, 0, true},
		{"invalid chars", "1,234.56", USD, 0, true},
		{"empty", "", USD, 0, true},
		{"no digits", ".", USD, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromMajorString(tt.input, tt.cur)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Minor != tt.want || got.Currency != tt.cur {
				t.Fatalf("got %v, want minor %d currency %v", got, tt.want, tt.cur)
			}
		})
	}
}

func TestFromMajorFloatRounding(t *testing.T) {
	a := FromMajorFloat(1.2345, USD) // 123.45 -> rounds to 123
	if a.Minor != 123 {
		t.Fatalf("got %d, want 123", a.Minor)
	}

	b := FromMajorFloat(-1.235, USD) // -123.5 -> rounds away from zero to -124
	if b.Minor != -124 {
		t.Fatalf("got %d, want -124", b.Minor)
	}
}

func TestMajorStringAndFloat(t *testing.T) {
	a := FromMinor(-5_002_344, USD)
	if got := a.MajorString(); got != "50023.44" {
		t.Fatalf("MajorString got %q, want %q", got, "50023.44")
	}
	if got := a.MajorFloat(); got != -50023.44 {
		t.Fatalf("MajorFloat got %f, want %f", got, -50023.44)
	}

	zeroScale := FromMinor(123, KRW)
	if got := zeroScale.MajorString(); got != "123" {
		t.Fatalf("MajorString scale 0 got %q, want %q", got, "123")
	}
}

func TestArithmetic(t *testing.T) {
	a := FromMinor(100, USD)
	b := FromMinor(50, USD)
	if got := a.Add(b); got.Minor != 150 {
		t.Fatalf("Add got %d, want 150", got.Minor)
	}
	if got := a.Sub(b); got.Minor != 50 {
		t.Fatalf("Sub got %d, want 50", got.Minor)
	}
	if got := a.MulInt(3); got.Minor != 300 {
		t.Fatalf("MulInt got %d, want 300", got.Minor)
	}
	if got := a.Neg(); got.Minor != -100 {
		t.Fatalf("Neg got %d, want -100", got.Minor)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on mismatched currency")
		}
	}()
	_ = a.Add(FromMinor(10, EUR))
}

func TestNegOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on overflow")
		}
	}()
	_ = FromMinor(math.MinInt64, USD).Neg()
}

func TestFormatterStandard(t *testing.T) {
	amt := FromMinor(5_002_344, USD)
	f := US()
	view := f.View(amt)

	if view.Formatted != "$50,023.44" || view.Sign != "" {
		t.Fatalf("formatted/sign got %q %q, want %q %q", view.Formatted, view.Sign, "$50,023.44", "")
	}
	if view.Major != "50023.44" || view.Minor != 5_002_344 || view.Currency != "USD" {
		t.Fatalf("view data mismatch: %+v", view)
	}

	custom := NewFormatter(FormatterConfig{
		Style:       StyleCurrency,
		UseGrouping: false,
		ShowPlus:    true,
	})
	if got := custom.Format(amt); got != "+$50023.44" {
		t.Fatalf("custom format got %q, want %q", got, "+$50023.44")
	}
	if got := custom.View(amt).Sign; got != "+" {
		t.Fatalf("sign got %q, want %q", got, "+")
	}
}

func TestFormatterAccountingAndZero(t *testing.T) {
	neg := FromMinor(-12_345, USD)
	ac := USAccounting()
	if got := ac.Format(neg); got != "($123.45)" {
		t.Fatalf("accounting format got %q, want %q", got, "($123.45)")
	}
	if sign := ac.View(neg).Sign; sign != "-" {
		t.Fatalf("accounting sign got %q, want %q", sign, "-")
	}

	zero := FromMinor(0, USD)
	withZero := NewFormatter(FormatterConfig{ZeroText: "—"})
	if got := withZero.Format(zero); got != "—" {
		t.Fatalf("zero text got %q, want %q", got, "—")
	}
	if sign := withZero.View(zero).Sign; sign != "" {
		t.Fatalf("zero sign got %q, want empty", sign)
	}
}

func TestFormatterISO(t *testing.T) {
	amt := FromMinor(5_002_344, USD)
	iso := NewFormatter(FormatterConfig{
		Style:       StyleISO,
		UseGrouping: true,
	})
	if got := iso.Format(amt); got != "USD 50,023.44" {
		t.Fatalf("ISO format got %q, want %q", got, "USD 50,023.44")
	}
}

func TestFormatCompact(t *testing.T) {
	large := FromMinor(123_456_789, USD) // 1,234,567.89
	if got := US().FormatCompact(large); got != "$1.2M" {
		t.Fatalf("compact got %q, want %q", got, "$1.2M")
	}

	neg := FromMinor(-150_000, USD) // -1,500.00
	if got := US().FormatCompact(neg); got != "-$1.5k" {
		t.Fatalf("compact negative got %q, want %q", got, "-$1.5k")
	}

	small := FromMinor(999_00, USD) // 999.00
	if got := US().FormatCompact(small); got != "$999.00" {
		t.Fatalf("compact small got %q, want %q", got, "$999.00")
	}

	roundUp := FromMinor(99_995_000, USD) // 999,950.00 -> should round to 1M
	if got := US().FormatCompact(roundUp); got != "$1M" {
		t.Fatalf("compact round-up got %q, want %q", got, "$1M")
	}
}

func TestMajorStringMinInt64(t *testing.T) {
	a := FromMinor(math.MinInt64, Currency{Code: "X", Symbol: "X", Scale: 0})
	if s := a.MajorString(); s != "9223372036854775808" {
		t.Fatalf("MajorString min int got %q", s)
	}
}
