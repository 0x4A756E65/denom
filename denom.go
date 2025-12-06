package denom

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Currency describes a fiat, crypto, or token with a business precision scale.
type Currency struct {
	Code   string
	Symbol string
	Scale  int
}

// Predefined fiat currencies.
var (
	USD = Currency{Code: "USD", Symbol: "$", Scale: 2}
	EUR = Currency{Code: "EUR", Symbol: "€", Scale: 2}
	GBP = Currency{Code: "GBP", Symbol: "£", Scale: 2}
	CHF = Currency{Code: "CHF", Symbol: "CHF", Scale: 2}
	CAD = Currency{Code: "CAD", Symbol: "CA$", Scale: 2}

	JPY = Currency{Code: "JPY", Symbol: "¥", Scale: 0}
	CNY = Currency{Code: "CNY", Symbol: "¥", Scale: 2}
	AUD = Currency{Code: "AUD", Symbol: "A$", Scale: 2}
	NZD = Currency{Code: "NZD", Symbol: "NZ$", Scale: 2}
	SGD = Currency{Code: "SGD", Symbol: "S$", Scale: 2}

	SEK = Currency{Code: "SEK", Symbol: "kr", Scale: 2}
	NOK = Currency{Code: "NOK", Symbol: "kr", Scale: 2}
	MXN = Currency{Code: "MXN", Symbol: "MX$", Scale: 2}
	BRL = Currency{Code: "BRL", Symbol: "R$", Scale: 2}
	INR = Currency{Code: "INR", Symbol: "₹", Scale: 2}
	HKD = Currency{Code: "HKD", Symbol: "HK$", Scale: 2}
	KRW = Currency{Code: "KRW", Symbol: "₩", Scale: 0}
	TRY = Currency{Code: "TRY", Symbol: "₺", Scale: 2}
	ZAR = Currency{Code: "ZAR", Symbol: "R", Scale: 2}
	PLN = Currency{Code: "PLN", Symbol: "zł", Scale: 2}
	RUB = Currency{Code: "RUB", Symbol: "₽", Scale: 2}
)

// Predefined crypto majors.
var (
	BTC = Currency{Code: "BTC", Symbol: "₿", Scale: 8}
	ETH = Currency{Code: "ETH", Symbol: "Ξ", Scale: 6}
)

// Predefined USD stablecoins (business precision).
var (
	USDC = Currency{Code: "USDC", Symbol: "USDC", Scale: 6}
	USDT = Currency{Code: "USDT", Symbol: "USDT", Scale: 6}
)

// Registry is a helper map for code-based lookups.
var Registry = map[string]Currency{
	"USD": USD, "EUR": EUR, "GBP": GBP,
	"CHF": CHF, "CAD": CAD, "JPY": JPY,
	"CNY": CNY, "AUD": AUD, "NZD": NZD,
	"SGD": SGD, "SEK": SEK, "NOK": NOK,
	"MXN": MXN, "BRL": BRL, "INR": INR,
	"HKD": HKD, "KRW": KRW, "TRY": TRY,
	"ZAR": ZAR, "PLN": PLN, "RUB": RUB,
	"BTC": BTC, "ETH": ETH,
	"USDC": USDC, "USDT": USDT,
}

// Lookup returns a currency from the registry by code.
func Lookup(code string) (Currency, bool) {
	c, ok := Registry[code]
	return c, ok
}

// Amount represents a value in a specific currency.
type Amount struct {
	Minor    int64
	Currency Currency
}

// FromMinor constructs an Amount from a minor-unit integer.
func FromMinor(minor int64, c Currency) Amount {
	return Amount{Minor: minor, Currency: c}
}

// FromMajorString parses a major-unit decimal string into an Amount.
// The string may start with an optional sign and must contain at most
// c.Scale fractional digits (no grouping separators).
func FromMajorString(s string, c Currency) (Amount, error) {
	pow, ok := pow10(c.Scale)
	if !ok {
		return Amount{}, fmt.Errorf("denom: unsupported scale %d", c.Scale)
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return Amount{}, errors.New("denom: empty major string")
	}

	sign := int64(1)
	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		sign = -1
		s = s[1:]
	}

	if s == "" {
		return Amount{}, errors.New("denom: invalid major string")
	}

	if !hasDigit(s) {
		return Amount{}, errors.New("denom: invalid major string")
	}

	if strings.Count(s, ".") > 1 {
		return Amount{}, errors.New("denom: invalid major string")
	}

	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}
	if intPart == "" {
		intPart = "0"
	}

	if !allDigits(intPart) || !allDigits(fracPart) {
		return Amount{}, errors.New("denom: invalid major string")
	}

	if len(fracPart) > c.Scale {
		return Amount{}, fmt.Errorf("denom: too many fractional digits for %s", c.Code)
	}

	fracPartPadded := fracPart + strings.Repeat("0", c.Scale-len(fracPart))

	intVal, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return Amount{}, fmt.Errorf("denom: %w", err)
	}

	intMinor, ok := checkedMul(intVal, pow)
	if !ok {
		return Amount{}, errors.New("denom: amount overflow")
	}

	var fracMinor int64
	if fracPartPadded != "" {
		fracMinor, err = strconv.ParseInt(fracPartPadded, 10, 64)
		if err != nil {
			return Amount{}, fmt.Errorf("denom: %w", err)
		}
	}

	minor, ok := checkedAdd(intMinor, fracMinor)
	if !ok {
		return Amount{}, errors.New("denom: amount overflow")
	}

	minor, ok = checkedMul(minor, sign)
	if !ok {
		return Amount{}, errors.New("denom: amount overflow")
	}

	return Amount{Minor: minor, Currency: c}, nil
}

// FromMajorFloat is a convenience constructor that rounds half away from zero
// to c.Scale decimal places.
func FromMajorFloat(f float64, c Currency) Amount {
	pow, ok := pow10(c.Scale)
	if !ok {
		panic(fmt.Sprintf("denom: unsupported scale %d", c.Scale))
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		panic("denom: invalid float")
	}

	minorF := math.Round(f * float64(pow))
	if minorF > float64(math.MaxInt64) || minorF < float64(math.MinInt64) {
		panic("denom: amount overflow")
	}

	return Amount{Minor: int64(minorF), Currency: c}
}

// IsZero reports whether the amount is zero.
func (a Amount) IsZero() bool {
	return a.Minor == 0
}

// IsNegative reports whether the amount is negative.
func (a Amount) IsNegative() bool {
	return a.Minor < 0
}

// MinorInt returns the raw minor-unit integer.
func (a Amount) MinorInt() int64 {
	return a.Minor
}

// MajorString returns a normalized signless decimal string using the
// currency scale.
func (a Amount) MajorString() string {
	pow := mustPow10(a.Currency.Scale)
	absMinor := absInt64(a.Minor)

	if pow == 1 {
		return strconv.FormatUint(absMinor, 10)
	}

	intPart := absMinor / uint64(pow)
	fracPart := absMinor % uint64(pow)
	fracStr := fmt.Sprintf("%0*d", a.Currency.Scale, fracPart)

	return strconv.FormatUint(intPart, 10) + "." + fracStr
}

// MajorFloat returns a best-effort float64 representation.
func (a Amount) MajorFloat() float64 {
	pow := mustPow10(a.Currency.Scale)
	return float64(a.Minor) / float64(pow)
}

// SameCurrency reports whether a and b share the same currency code and scale.
func (a Amount) SameCurrency(b Amount) bool {
	return a.Currency.Code == b.Currency.Code && a.Currency.Scale == b.Currency.Scale
}

// Add adds two Amounts; panics if currencies differ or on overflow.
func (a Amount) Add(b Amount) Amount {
	if !a.SameCurrency(b) {
		panic("denom: currency mismatch")
	}
	minor, ok := checkedAdd(a.Minor, b.Minor)
	if !ok {
		panic("denom: amount overflow")
	}
	return Amount{Minor: minor, Currency: a.Currency}
}

// Sub subtracts b from a; panics if currencies differ or on overflow.
func (a Amount) Sub(b Amount) Amount {
	if !a.SameCurrency(b) {
		panic("denom: currency mismatch")
	}
	minor, ok := checkedSub(a.Minor, b.Minor)
	if !ok {
		panic("denom: amount overflow")
	}
	return Amount{Minor: minor, Currency: a.Currency}
}

// MulInt multiplies an Amount by an integer factor; panics on overflow.
func (a Amount) MulInt(n int64) Amount {
	minor, ok := checkedMul(a.Minor, n)
	if !ok {
		panic("denom: amount overflow")
	}
	return Amount{Minor: minor, Currency: a.Currency}
}

// Neg returns the additive inverse of the amount; panics on overflow.
func (a Amount) Neg() Amount {
	minor, ok := checkedMul(a.Minor, -1)
	if !ok {
		panic("denom: amount overflow")
	}
	return Amount{Minor: minor, Currency: a.Currency}
}

// View represents a formatted perspective on an amount.
type View struct {
	Minor int64   `json:"minor"`
	Major string  `json:"major"`
	Float float64 `json:"float,omitempty"`

	Formatted string `json:"formatted"`
	Sign      string `json:"sign,omitempty"`

	Currency       string `json:"currency"`
	CurrencySymbol string `json:"currency_symbol"`
}

// Style controls how formatted amounts are rendered.
type Style int

const (
	StyleCurrency Style = iota
	StyleNoSymbol
	StyleISO
	StyleAccounting
)

// FormatterConfig controls formatting behavior.
type FormatterConfig struct {
	Style       Style
	UseGrouping bool
	ShowPlus    bool
	ZeroText    string
	Locale      string
}

// Formatter renders amounts according to the config.
type Formatter struct {
	cfg FormatterConfig
}

var compactSuffixes = []string{"", "k", "M", "B", "T"}

// NewFormatter creates a formatter from the given config.
func NewFormatter(cfg FormatterConfig) Formatter {
	return Formatter{cfg: cfg}
}

// US returns a default US-style formatter.
func US() Formatter {
	return NewFormatter(FormatterConfig{
		Style:       StyleCurrency,
		UseGrouping: true,
		ShowPlus:    false,
		ZeroText:    "",
		Locale:      "en",
	})
}

// USAccounting returns a formatter with accounting negatives.
func USAccounting() Formatter {
	return NewFormatter(FormatterConfig{
		Style:       StyleAccounting,
		UseGrouping: true,
		ShowPlus:    false,
		ZeroText:    "",
		Locale:      "en",
	})
}

// View converts an Amount into a rich View struct.
func (f Formatter) View(a Amount) View {
	formatted, sign := f.formatStandard(a)
	return View{
		Minor:          a.Minor,
		Major:          a.MajorString(),
		Float:          a.MajorFloat(),
		Formatted:      formatted,
		Sign:           sign,
		Currency:       a.Currency.Code,
		CurrencySymbol: a.Currency.Symbol,
	}
}

// Format returns a human-friendly string for the amount.
func (f Formatter) Format(a Amount) string {
	formatted, _ := f.formatStandard(a)
	return formatted
}

// FormatCompact returns a compact representation like "$50k" or "$1.2M".
func (f Formatter) FormatCompact(a Amount) string {
	if f.cfg.ZeroText != "" && a.IsZero() {
		return f.cfg.ZeroText
	}

	pow := mustPow10(a.Currency.Scale)
	absMinor := absInt64(a.Minor)
	powU := uint64(pow)

	majorInt := absMinor / powU
	if majorInt < 1000 {
		return f.Format(a)
	}

	idx := 0
	for majorInt >= 1000 && idx < len(compactSuffixes)-1 {
		majorInt /= 1000
		idx++
	}

	divisor := float64(pow)
	if idx > 0 {
		divisor *= math.Pow(1000, float64(idx))
	}

	value := float64(absMinor) / divisor
	rounded := math.Round(value*10) / 10
	if rounded >= 1000 && idx < len(compactSuffixes)-1 {
		rounded /= 1000
		idx++
	}

	number := strconv.FormatFloat(rounded, 'f', 1, 64)
	number = strings.TrimSuffix(strings.TrimSuffix(number, "0"), ".")
	number += compactSuffixes[idx]

	formatted, _ := f.applyStyle(a, number)
	return formatted
}

func (f Formatter) formatStandard(a Amount) (string, string) {
	if f.cfg.ZeroText != "" && a.IsZero() {
		return f.cfg.ZeroText, ""
	}
	number := a.MajorString()
	if f.cfg.UseGrouping {
		number = addGrouping(number)
	}
	return f.applyStyle(a, number)
}

func (f Formatter) applyStyle(a Amount, number string) (string, string) {
	sign := ""
	if a.IsNegative() {
		sign = "-"
	} else if f.cfg.ShowPlus && !a.IsZero() {
		sign = "+"
	}

	switch f.cfg.Style {
	case StyleNoSymbol:
		return sign + number, sign
	case StyleISO:
		return sign + a.Currency.Code + " " + number, sign
	case StyleAccounting:
		if a.IsNegative() {
			return "(" + a.Currency.Symbol + number + ")", "-"
		}
		return sign + a.Currency.Symbol + number, sign
	default:
		return sign + a.Currency.Symbol + number, sign
	}
}

func pow10(scale int) (int64, bool) {
	if scale < 0 || scale > 18 {
		return 0, false
	}
	value := int64(1)
	for i := 0; i < scale; i++ {
		value *= 10
	}
	return value, true
}

func mustPow10(scale int) int64 {
	value, ok := pow10(scale)
	if !ok {
		panic(fmt.Sprintf("denom: unsupported scale %d", scale))
	}
	return value
}

func addGrouping(major string) string {
	dot := strings.IndexByte(major, '.')
	intPart := major
	fracPart := ""
	if dot >= 0 {
		intPart = major[:dot]
		fracPart = major[dot+1:]
	}

	if len(intPart) <= 3 {
		return major
	}

	start := len(intPart) % 3
	if start == 0 {
		start = 3
	}

	var b strings.Builder
	b.Grow(len(major) + len(intPart)/3)
	b.WriteString(intPart[:start])
	for i := start; i < len(intPart); i += 3 {
		b.WriteByte(',')
		b.WriteString(intPart[i : i+3])
	}

	if fracPart != "" {
		b.WriteByte('.')
		b.WriteString(fracPart)
	}

	return b.String()
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func hasDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			return true
		}
	}
	return false
}

func absInt64(v int64) uint64 {
	if v >= 0 {
		return uint64(v)
	}
	return uint64(-(v + 1) + 1)
}

func checkedAdd(a, b int64) (int64, bool) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, false
	}
	return a + b, true
}

func checkedSub(a, b int64) (int64, bool) {
	if (b > 0 && a < math.MinInt64+b) || (b < 0 && a > math.MaxInt64+b) {
		return 0, false
	}
	return a - b, true
}

func checkedMul(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	if (a == math.MinInt64 && b == -1) || (b == math.MinInt64 && a == -1) {
		return 0, false
	}
	res := a * b
	if res/b != a {
		return 0, false
	}
	return res, true
}
