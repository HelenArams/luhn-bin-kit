package bin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrNonDigit is returned by Validate when the number is empty or contains
// anything other than ASCII digits. Strip separators before calling, for
// example with the mask package's Normalize.
var ErrNonDigit = errors.New("bin: number must contain only digits")

// LengthError is returned by Validate when the number belongs to a known
// scheme but its length is not one that scheme issues.
type LengthError struct {
	Scheme Scheme
	Length int
}

func (e *LengthError) Error() string {
	return fmt.Sprintf("bin: %s numbers have %s digits, got %d",
		e.Scheme, describeLengths(e.Scheme.Lengths()), e.Length)
}

// Detect returns the scheme that owns the leading digits of pan. The input
// may be a complete number or only its first digits (an IIN/BIN) and must
// consist of ASCII digits; anything else yields Unknown and false.
//
// Detect looks at prefixes only. It does not check the length or the Luhn
// digit; see Validate and the luhn package for those.
func Detect(pan string) (Scheme, bool) {
	if !isDigits(pan) {
		return Unknown, false
	}
	for _, e := range table {
		if e.Contains(pan) {
			return e.scheme, true
		}
	}
	return Unknown, false
}

// Validate checks that pan is all digits, belongs to a known scheme and has
// a length that scheme issues. It returns the detected scheme together with
// a nil error, ErrNonDigit, ErrUnknownScheme or a *LengthError. When the
// error is a *LengthError the scheme is still returned so callers can report
// it.
//
// Validate does not verify the check digit: combine it with luhn.Valid.
func Validate(pan string) (Scheme, error) {
	if !isDigits(pan) {
		return Unknown, ErrNonDigit
	}
	s, ok := Detect(pan)
	if !ok {
		return Unknown, ErrUnknownScheme
	}
	if !s.ValidLength(len(pan)) {
		return s, &LengthError{Scheme: s, Length: len(pan)}
	}
	return s, nil
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

// describeLengths renders an ascending list of lengths for an error
// message: "15", "16 or 19", "13, 16 or 19", or "14 to 19" for a run of
// consecutive values.
func describeLengths(ls []int) string {
	switch len(ls) {
	case 0:
		return "no"
	case 1:
		return strconv.Itoa(ls[0])
	}
	if ls[len(ls)-1]-ls[0] == len(ls)-1 && len(ls) > 2 {
		return strconv.Itoa(ls[0]) + " to " + strconv.Itoa(ls[len(ls)-1])
	}
	parts := make([]string, len(ls))
	for i, l := range ls {
		parts[i] = strconv.Itoa(l)
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " or " + parts[len(parts)-1]
}
