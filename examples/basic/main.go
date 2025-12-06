package main

import (
	"fmt"

	denom "github.com/0x4A756E65"
)

func main() {
	// Build an amount from minor units.
	amt := denom.FromMinor(5_002_344, denom.USD) // 50,023.44 USD
	us := denom.US()
	view := us.View(amt)

	fmt.Println("Standard:")
	fmt.Printf("  formatted: %s\n", view.Formatted)
	fmt.Printf("  major:     %s\n", view.Major)
	fmt.Printf("  minor:     %d\n", view.Minor)
	fmt.Printf("  currency:  %s (%s)\n", view.Currency, view.CurrencySymbol)

	// Parse from a major string with scale validation.
	usdc, err := denom.FromMajorString("1234.567890", denom.USDC) // scale 6
	if err != nil {
		panic(err)
	}

	iso := denom.NewFormatter(denom.FormatterConfig{
		Style:       denom.StyleISO,
		UseGrouping: true,
	})

	fmt.Println("\nParsed (USDC scale 6):")
	fmt.Printf("  formatted: %s\n", iso.Format(usdc))
	fmt.Printf("  major:     %s\n", usdc.MajorString())
	fmt.Printf("  compact:   %s\n", iso.FormatCompact(usdc))

	// Accounting style for negatives.
	loss := denom.FromMinor(-12_345, denom.USD)
	acct := denom.USAccounting()
	fmt.Println("\nAccounting style:")
	fmt.Printf("  formatted: %s\n", acct.Format(loss))
	fmt.Printf("  sign:      %q\n", acct.View(loss).Sign)
}
