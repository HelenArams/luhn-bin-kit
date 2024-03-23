package mask

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// MinLength is the shortest digit string Mask accepts. Shorter inputs
	// are not card numbers and are rejected with ErrTooShort rather than
	// being masked in a way that would reveal most of them.
	MinLength = 8

	// Char is the character Mask substitutes for hidden digits.
	Char = '*'

	keepPrefix = 6
	keepSuffix = 4
)

// ErrEmpty is returned when the input holds no digits at all.
var ErrEmpty = errors.New("mask: empty input")

// ErrTooShort is returned when the input has too few digits for the
// requested operation: fewer than MinLength for Mask, fewer than four for
// Last4.
var ErrTooShort = errors.New("mask: number too short")

// InvalidCharError reports the first character of the input that is neither
// a digit nor an accepted separator. Pos is the byte offset in the input.
type InvalidCharError struct {
	Pos  int
	Char rune
}

func (e *InvalidCharError) Error() string {
	return fmt.Sprintf("mask: invalid character %q at position %d", e.Char, e.Pos)
}

// Normalize strips spaces and hyphens from s and returns the remaining ASCII
// digits. Any other character, including tabs and non-ASCII digits, yields
// an *InvalidCharError; an input with no digits yields ErrEmpty. When s is
// already a digit string it is returned as is without allocating.
func Normalize(s string) (string, error) {
	if isDigits(s) {
		return s, nil
	}
	var b strings.Builder
	b.Grow(len(s))
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteByte(byte(r))
		case r == ' ' || r == '-':
			continue
		default:
			return "", &InvalidCharError{Pos: i, Char: r}
		}
	}
	if b.Len() == 0 {
		return "", ErrEmpty
	}
	return b.String(), nil
}

// Mask normalizes pan and hides every digit except the first six and the
// last four, so "4111 1111 1111 1111" becomes "411111******1111". The result
// always has one character per digit of the normalized input.
//
// Inputs shorter than MinLength digits are rejected with ErrTooShort; the
// errors of Normalize are passed through.
func Mask(pan string) (string, error) {
	d, err := Normalize(pan)
	if err != nil {
		return "", err
	}
	if len(d) < MinLength {
		return "", ErrTooShort
	}
	// PCI allows first 6 + last 4, but on a short number that would show
	// everything, so below 10 digits only the last four survive.
	keep := keepPrefix
	if len(d) < keepPrefix+keepSuffix {
		keep = 0
	}
	return hide(d, keep, keepSuffix), nil
}

// Last4 normalizes pan and returns its last four digits, the form most
// often stored alongside a token. Inputs with fewer than four digits yield
// ErrTooShort; the errors of Normalize are passed through.
func Last4(pan string) (string, error) {
	d, err := Normalize(pan)
	if err != nil {
		return "", err
	}
	if len(d) < keepSuffix {
		return "", ErrTooShort
	}
	return d[len(d)-keepSuffix:], nil
}

func hide(d string, prefix, suffix int) string {
	b := make([]byte, len(d))
	copy(b, d[:prefix])
	for i := prefix; i < len(d)-suffix; i++ {
		b[i] = Char
	}
	copy(b[len(d)-suffix:], d[len(d)-suffix:])
	return string(b)
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
