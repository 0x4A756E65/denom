package main

import (
	"encoding/json"
	"fmt"

	denom "github.com/0x4A756E65"
)

func main() {
	// ISO formatting with optional plus sign.
	eur, ok := denom.Lookup("EUR")
	if !ok {
		panic("EUR missing")
	}
	salary, err := denom.FromMajorString("4321.09", eur)
	if err != nil {
		panic(err)
	}
	bonus := denom.FromMinor(7_500, eur) // €75.00
	total := salary.Add(bonus)

	isoPlus := denom.NewFormatter(denom.FormatterConfig{
		Style:       denom.StyleISO,
		UseGrouping: true,
		ShowPlus:    true,
	})
	fmt.Println("ISO with plus:", isoPlus.Format(total)) // "+EUR 4,396.09"

	// Accounting style for negatives.
	loss := denom.FromMajorFloat(-1234.56, denom.USD)
	fmt.Println("Accounting:", denom.USAccounting().Format(loss)) // "($1,234.56)"

	// Compact formatting for crypto using ISO style.
	eth := denom.FromMinor(12_345_678_000, denom.ETH) // 12,345.678000 ETH (scale 6)
	iso := denom.NewFormatter(denom.FormatterConfig{
		Style:       denom.StyleISO,
		UseGrouping: true,
	})
	fmt.Println("ETH compact:", iso.FormatCompact(eth)) // "ETH 12.3k"

	// Custom currency with scale 3 and no grouping.
	TOK := denom.Currency{Code: "TOK", Symbol: "¤", Scale: 3}
	tokAmt, err := denom.FromMajorString("10.123", TOK)
	if err != nil {
		panic(err)
	}
	custom := denom.NewFormatter(denom.FormatterConfig{
		Style:       denom.StyleCurrency,
		UseGrouping: false,
		ZeroText:    "—",
	})
	view := custom.View(tokAmt)
	buf, _ := json.MarshalIndent(view, "", "  ")
	fmt.Println("Custom currency JSON view:")
	fmt.Println(string(buf))
}
