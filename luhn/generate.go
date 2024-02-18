package luhn

import (
	"errors"
	"math/rand"
)

// ErrLength is returned by Generate when the requested length leaves no room
// for a check digit after the prefix.
var ErrLength = errors.New("luhn: length must be greater than the prefix length")

// Source is the randomness Generate draws from. *math/rand.Rand satisfies it,
// and so does any type that can hand out a uniform integer in [0, n).
type Source interface {
	Intn(n int) int
}

// stdSource adapts the auto-seeded package-level math/rand functions.
type stdSource struct{}

func (stdSource) Intn(n int) int { return rand.Intn(n) } //nolint:gosec // test data, not secrets

// Generate returns a number of exactly length digits that starts with prefix
// and ends with a valid Luhn check digit. The digits between the prefix and
// the check digit are drawn from src; when src is nil the package-level
// math/rand source is used.
//
// The output is suitable for test fixtures only: it is not a real card
// number and carries no guarantee about which scheme, if any, owns the
// prefix. Prefix must consist of ASCII digits (it may be empty) and length
// must exceed len(prefix), otherwise a *NonDigitError or ErrLength is
// returned.
func Generate(prefix string, length int, src Source) (string, error) {
	if _, bad := sum(prefix, 0); bad >= 0 {
		return "", &NonDigitError{Pos: bad, Char: prefix[bad]}
	}
	if length <= len(prefix) {
		return "", ErrLength
	}
	if src == nil {
		src = stdSource{}
	}

	buf := make([]byte, length)
	copy(buf, prefix)
	for i := len(prefix); i < length-1; i++ {
		buf[i] = '0' + byte(src.Intn(10))
	}

	// The payload is all digits by construction, so sum cannot fail here.
	total, _ := sum(string(buf[:length-1]), 1)
	buf[length-1] = digitFor(total)
	return string(buf), nil
}
