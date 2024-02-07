package luhn

import (
	"errors"
	"fmt"
)

// ErrEmpty is returned when an operation receives an empty payload.
var ErrEmpty = errors.New("luhn: empty input")

// NonDigitError reports the first byte of the input that is not an ASCII
// digit. Pos is the byte offset into the original string.
type NonDigitError struct {
	Pos  int
	Char byte
}

func (e *NonDigitError) Error() string {
	return fmt.Sprintf("luhn: non-digit character %q at position %d", e.Char, e.Pos)
}

// doubled maps a digit to the value it contributes when it sits in a
// doubled position: 2*d, minus 9 when the product has two digits.
var doubled = [10]int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}

// Valid reports whether s is a string of at least two ASCII digits whose
// last digit is a correct Luhn check digit. Anything else, including an
// empty string, separators or a single digit, is reported as invalid.
func Valid(s string) bool {
	if len(s) < 2 {
		return false
	}
	total, bad := sum(s, 0)
	return bad < 0 && total%10 == 0
}

// CheckDigit computes the Luhn check digit for payload, the number without
// its final digit, and returns it as an ASCII digit ('0'..'9').
//
// It returns ErrEmpty for an empty payload and a *NonDigitError when the
// payload contains anything other than ASCII digits.
func CheckDigit(payload string) (byte, error) {
	if payload == "" {
		return 0, ErrEmpty
	}
	total, bad := sum(payload, 1)
	if bad >= 0 {
		return 0, &NonDigitError{Pos: bad, Char: payload[bad]}
	}
	return '0' + byte((10-total%10)%10), nil
}

// sum folds the Luhn sum over s. Digits are doubled based on their distance
// from the right end: with parity 0 the rightmost digit is left alone, which
// is the rule for a complete number; with parity 1 the rightmost digit is
// doubled, which is the rule for a payload whose check digit is still to be
// appended. The second result is the index of the first non-digit byte, or
// -1 when s is all digits.
func sum(s string, parity int) (total int, bad int) {
	n := len(s)
	for i := 0; i < n; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, i
		}
		d := int(c - '0')
		if (n-i)%2 == parity {
			d = doubled[d]
		}
		total += d
	}
	return total, -1
}
