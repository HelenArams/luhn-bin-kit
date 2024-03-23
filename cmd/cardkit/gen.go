package main

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"time"

	"github.com/HelenArams/luhn-bin-kit/bin"
	"github.com/HelenArams/luhn-bin-kit/luhn"
)

const maxGenCount = 10000

func cmdGen(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("gen", "--scheme <name> [-count n] [-length n] [-seed n]", stderr)
	schemeName := fs.String("scheme", "visa", "card scheme to generate numbers for (visa, mastercard, amex, ...)")
	count := fs.Int("count", 1, "how many numbers to print")
	length := fs.Int("length", 0, "number length; defaults to the scheme's most common length")
	seed := fs.Int64("seed", 0, "random seed for reproducible output; 0 picks one from the clock")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "cardkit gen: unexpected argument %q\n", fs.Arg(0))
		return exitUsage
	}

	scheme, err := bin.ParseScheme(*schemeName)
	if err != nil || scheme == bin.Unknown {
		fmt.Fprintf(stderr, "cardkit gen: unknown scheme %q\n", *schemeName)
		return exitUsage
	}
	if *count < 1 || *count > maxGenCount {
		fmt.Fprintf(stderr, "cardkit gen: -count must be between 1 and %d\n", maxGenCount)
		return exitUsage
	}
	n := *length
	if n == 0 {
		n = defaultLength(scheme)
	}
	if !scheme.ValidLength(n) {
		fmt.Fprintf(stderr, "cardkit gen: %s numbers can't be %d digits (want one of %v)\n", scheme, n, scheme.Lengths())
		return exitUsage
	}

	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(s)) //nolint:gosec // test data, not secrets

	ranges := bin.Ranges(scheme)
	for i := 0; i < *count; i++ {
		fmt.Fprintln(stdout, generate(scheme, ranges, n, rng))
	}
	return exitOK
}

// generate keeps drawing until Detect agrees with the requested scheme.
// Short ranges can contain another scheme's carve-out (UnionPay's 62 holds
// Discover's 622126-622925), so a random fill can land in the wrong place.
// It's rare enough that retrying is simpler than excluding sub-ranges.
func generate(scheme bin.Scheme, ranges []bin.Range, n int, rng *rand.Rand) string {
	for {
		r := ranges[rng.Intn(len(ranges))]
		pan, err := luhn.Generate(pickPrefix(r, rng), n, rng)
		if err != nil {
			// prefixes are digits and n is a valid scheme length (>= 12),
			// so this would be a bug in the table
			panic(err)
		}
		if got, _ := bin.Detect(pan); got == scheme {
			return pan
		}
	}
}

func defaultLength(s bin.Scheme) int {
	ls := s.Lengths()
	for _, l := range ls {
		if l == 16 {
			return 16
		}
	}
	return ls[0]
}

func pickPrefix(r bin.Range, rng *rand.Rand) string {
	lo, _ := strconv.Atoi(r.Lo)
	hi, _ := strconv.Atoi(r.Hi)
	return fmt.Sprintf("%0*d", r.Digits(), lo+rng.Intn(hi-lo+1))
}
