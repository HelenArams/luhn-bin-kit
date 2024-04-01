# luhn-bin-kit

[![ci](https://github.com/HelenArams/luhn-bin-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/HelenArams/luhn-bin-kit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/HelenArams/luhn-bin-kit.svg)](https://pkg.go.dev/github.com/HelenArams/luhn-bin-kit)

I got tired of pasting test PANs into random online Luhn checkers (and of
re-writing the same `isValidLuhn` helper in every service that touches card
numbers), so I put the bits I keep needing into one small Go module:

- `luhn` - check digit validation, computing a check digit, generating test numbers
- `bin` - which scheme owns a number (or just a BIN), and whether the length makes sense for it
- `mask` - `411111******1111` / last four, tolerant of spaces and dashes
- `cardkit` - a CLI on top of all three, because half the time I just want an answer in the terminal

No dependencies outside the standard library. MIT licensed.

It is deliberately small. It doesn't know issuers, countries or card
products - for that you need a real BIN database, not a table in a Go file.

## Install

```sh
go get github.com/HelenArams/luhn-bin-kit
go install github.com/HelenArams/luhn-bin-kit/cmd/cardkit@latest
```

Needs Go 1.21+.

## Using the library

```go
import (
	"github.com/HelenArams/luhn-bin-kit/bin"
	"github.com/HelenArams/luhn-bin-kit/luhn"
	"github.com/HelenArams/luhn-bin-kit/mask"
)

pan, err := mask.Normalize("4111 1111 1111 1111") // "4111111111111111"
if err != nil {
	// *mask.InvalidCharError tells you which character and where
}

luhn.Valid(pan) // true

scheme, err := bin.Validate(pan) // bin.Visa, nil
// errors: bin.ErrNonDigit, bin.ErrUnknownScheme or *bin.LengthError

bin.Detect("37")        // bin.Amex, true - a prefix is enough
bin.Amex.Lengths()      // [15]

masked, _ := mask.Mask(pan) // "411111******1111"
last4, _ := mask.Last4(pan) // "1111"
```

Generating numbers for fixtures:

```go
d, _ := luhn.CheckDigit("411111111111111") // '1'

rng := rand.New(rand.NewSource(1))     // or nil for math/rand's default source
n, _ := luhn.Generate("5555", 16, rng) // 16 digits, starts with 5555, passes Valid
```

A couple of choices worth knowing about:

- `luhn.Valid` does **not** strip separators. "4111 1111 1111 1111" is
  invalid. Normalize first - I'd rather it fail loudly than quietly accept
  whatever garbage came in.
- `bin.Scheme` implements `encoding.TextMarshaler`, so it goes into JSON as
  `"amex"`, `"visa"` etc. and `ParseScheme` understands `"American Express"`,
  `"mc"`, `"cup"` and friends.
- Every error is either a sentinel (`errors.Is`) or a typed error
  (`errors.As`), documented on the function that returns it.

## cardkit

```console
$ cardkit validate "4111 1111 1111 1111"
valid: Visa, 16 digits

$ cardkit validate 411111111111116
invalid: Visa numbers have 13, 16 or 19 digits, got 15

$ cardkit detect 6221260000000000 37 9999
622126******0000	Discover
37	American Express
9999	unknown

$ cardkit mask 4111-1111-1111-1111
411111******1111

$ cardkit gen --scheme visa --count 3 --seed 1
4791850604129844
4765688777805184
4967156304337846
```

Without arguments it reads numbers from stdin, one per line, so you can
pipe a file through it. When there's more than one number the output is
labelled with the masked input - it never echoes a full PAN back.

```console
$ cat pans.txt | cardkit validate
411111******1111	valid: Visa, 16 digits
601111******1117	valid: Discover, 16 digits
555555******4445	invalid: luhn check failed
```

Exit code is 0 if everything passed, 1 if anything failed, 2 for usage
errors. `cardkit <command> -h` lists the flags (`validate -strict`,
`detect -id`, `mask -last4`, `gen -length/-seed`).

> **The numbers `gen` (and `luhn.Generate`) produce are test data.** They
> pass Luhn and land in the right range, which is exactly what sandbox
> environments and unit tests want. They aren't real cards and won't work
> anywhere real.

## Schemes it knows

| Scheme           | Prefixes                                   | Lengths       |
|------------------|--------------------------------------------|---------------|
| Visa             | 4                                          | 13, 16, 19    |
| Mastercard       | 51-55, 2221-2720                           | 16            |
| American Express | 34, 37                                     | 15            |
| Discover         | 6011, 644-649, 65, 622126-622925           | 16, 19        |
| JCB              | 3528-3589                                  | 16-19         |
| UnionPay         | 62 (except Discover's 622126-622925)       | 16-19         |
| Maestro          | 50, 56-58, 6304, 67                        | 12-19         |
| Diners Club      | 300-305, 3095, 36, 38-39                   | 14-19         |

Longest prefix wins, which is how 622126 ends up as Discover rather than
UnionPay. A test checks the table is ordered that way and that ranges of
the same width don't overlap, so adding a range in the wrong place fails CI.

Ranges are from public scheme docs and they do move around - if you spot
one that's out of date, an issue or PR is very welcome.

## Layout

```
luhn/          checksum, check digit, Generate
bin/           Scheme, range table, Detect / Validate
mask/          Normalize, Mask, Last4
cmd/cardkit/   the CLI
```

## Tests

```sh
make test     # go test -race -cover ./...
make lint     # installs golangci-lint v1.56.2 into .tools/ on first run
make fuzz     # FuzzValid and FuzzMask, 30s each (FUZZTIME=5m make fuzz)
make bench
```

The `Example` functions double as tests, so the snippets in the godoc are
checked on every run. Coverage sits at ~100% for the library packages.

## Next

- [ ] 8-digit BIN ranges (the move happened in 2022, haven't found a clean public list to embed yet)
- [ ] `cardkit detect -json` for piping into jq
- [ ] maybe RuPay and Mir, if anyone asks

## License

MIT - see [LICENSE](LICENSE).
