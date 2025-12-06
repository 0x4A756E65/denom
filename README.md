# denom – Monetary Amount & Formatting for Go

![denom logo](assets/denom-logo.svg)

Small, dependency-free helpers for representing money with integers and formatting it consistently across APIs, logs, and UIs.

## Install

```bash
go get github.com/0x4A756E65
```

## Core Ideas

- **Amount** stores only minor units (`int64`) plus a **Currency** that defines its decimal **Scale** (business precision, e.g., 2 for USD cents, 0 for JPY, 6 for USDC).
- **Formatter** renders an Amount into a reusable **View** with precomputed strings: raw minor, normalized major string, optional float, formatted display, sign, and currency metadata.
- Arithmetic is integer-only; formatting is configurable (grouping, symbols, ISO/accounting style, optional plus, custom zero text).

## Quick Start

```go
amt := denom.FromMinor(5_002_344, denom.USD) // 50,023.44 USD
view := denom.US().View(amt)

fmt.Println(view.Formatted) // "$50,023.44"
fmt.Println(view.Major)     // "50023.44"
fmt.Println(view.Minor)     // 5002344
fmt.Println(view.Currency)  // "USD"
```

Parse from a string with scale-aware validation:

```go
amt, err := denom.FromMajorString("1234.567890", denom.USDC) // scale 6
if err != nil {
	log.Fatal(err)
}
iso := denom.NewFormatter(denom.FormatterConfig{Style: denom.StyleISO, UseGrouping: true})
fmt.Println(iso.Format(amt)) // "USDC 1,234.567890"
```

Use accounting style and compact display:

```go
loss := denom.FromMinor(-12_345, denom.USD) // -123.45
fmt.Println(denom.USAccounting().Format(loss))      // "($123.45)"
fmt.Println(denom.US().FormatCompact(loss.MulInt(10))) // "-$1.2k"
```

Run the example program:

```bash
go run ./examples/basic
```

## Formatting Styles

- `StyleCurrency` (default): `$50,023.44`
- `StyleNoSymbol`: `50,023.44`
- `StyleISO`: `USD 50,023.44`
- `StyleAccounting`: `($50,023.44)` for negatives

Flags: `UseGrouping` (commas), `ShowPlus` (e.g., `+$50.00`), `ZeroText` (e.g., `—`).

## Predefined Currencies

Fiat: USD, EUR, GBP, CHF, CAD, JPY, CNY, AUD, NZD, SGD, SEK, NOK, MXN, BRL, INR, HKD, KRW, TRY, ZAR, PLN, RUB.  
Crypto: BTC (scale 8), ETH (scale 6).  
Stablecoins: USDC (scale 6), USDT (scale 6).  
Use `denom.Lookup(code)` or define your own `Currency` with the scale your product uses.

## Development

```bash
go test ./...
```
