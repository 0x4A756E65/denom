package main

import (
	"encoding/json"
	"fmt"

	denom "github.com/0x4A756E65"
)

func main() {
	fmt.Println("== Basic ==")
	basic()

	fmt.Println("\n== Parsing and validation ==")
	parsing()

	fmt.Println("\n== Accounting style and plus sign ==")
	signs()

	fmt.Println("\n== Compact formatting ==")
	compact()

	fmt.Println("\n== Custom currency and zero text ==")
	custom()

	fmt.Println("\n== Registry lookup and arithmetic ==")
	registryAndMath()
}

func basic() {
	amt := denom.FromMinor(5_002_344, denom.USD) // 50,023.44 USD
	view := denom.US().View(amt)
	fmt.Printf("Formatted: %s | Major: %s | Minor: %d | Sign: %q | Currency: %s (%s)\n",
		view.Formatted, view.Major, view.Minor, view.Sign, view.Currency, view.CurrencySymbol)
}

func parsing() {
	usdc, err := denom.FromMajorString("1234.567890", denom.USDC) // scale 6
	if err != nil {
		panic(err)
	}
	iso := denom.NewFormatter(denom.FormatterConfig{Style: denom.StyleISO, UseGrouping: true})
	fmt.Printf("USDC parsed: %s (major=%s minor=%d)\n", iso.Format(usdc), usdc.MajorString(), usdc.MinorInt())

	_, err = denom.FromMajorString("1.234", denom.USD) // too many decimals for USD
	fmt.Printf("Parse error (expected): %v\n", err)
}

func signs() {
	isoPlus := denom.NewFormatter(denom.FormatterConfig{
		Style:       denom.StyleISO,
		UseGrouping: true,
		ShowPlus:    true,
	})
	eur, _ := denom.Lookup("EUR")
	val, _ := denom.FromMajorString("4321.09", eur)
	fmt.Println("ISO with plus:", isoPlus.Format(val))

	loss := denom.FromMajorFloat(-1234.56, denom.USD)
	fmt.Println("Accounting negative:", denom.USAccounting().Format(loss))
}

func compact() {
	eth := denom.FromMinor(12_345_678_000, denom.ETH) // 12,345.678000 ETH (scale 6)
	iso := denom.NewFormatter(denom.FormatterConfig{
		Style:       denom.StyleISO,
		UseGrouping: true,
	})
	fmt.Println("ETH compact:", iso.FormatCompact(eth))

	largeUSD := denom.FromMinor(99_995_000, denom.USD) // 999,950.00 -> 1M
	fmt.Println("USD compact round:", denom.US().FormatCompact(largeUSD))
}

func custom() {
	TOK := denom.Currency{Code: "TOK", Symbol: "¤", Scale: 3}
	tokAmt, _ := denom.FromMajorString("10.123", TOK)
	withZero := denom.NewFormatter(denom.FormatterConfig{
		Style:       denom.StyleCurrency,
		UseGrouping: false,
		ZeroText:    "—",
	})
	view := withZero.View(tokAmt)
	buf, _ := json.MarshalIndent(view, "", "  ")
	fmt.Println("Custom currency JSON:\n" + string(buf))

	zero := denom.FromMinor(0, TOK)
	fmt.Println("Zero text:", withZero.Format(zero))
}

func registryAndMath() {
	usd := denom.USD
	base := denom.FromMinor(10_000, usd) // $100.00
	fee := denom.FromMinor(125, usd)     // $1.25
	total := base.Sub(fee)
	fmt.Println("Total after fee:", denom.US().Format(total))

	// Same currency enforcement demonstrated with panic recovery.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Add mismatched currencies panicked as expected:", r)
		}
	}()
	_ = base.Add(denom.FromMinor(1, denom.EUR))
}
