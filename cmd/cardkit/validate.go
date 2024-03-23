package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/HelenArams/luhn-bin-kit/bin"
	"github.com/HelenArams/luhn-bin-kit/luhn"
	"github.com/HelenArams/luhn-bin-kit/mask"
)

// Numbers of an unknown scheme cannot be checked against a scheme's lengths
// but are still held to the ISO/IEC 7812 envelope.
const (
	minPANLength = 8
	maxPANLength = 19
)

func cmdValidate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("validate", "[-strict] [number...]", stderr)
	strict := fs.Bool("strict", false, "reject numbers whose scheme is not recognised")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	return each(fs.Args(), stdin, stdout, stderr, len(fs.Args()) != 1, func(in string) (string, bool) {
		return validate(in, *strict)
	})
}

// validate checks one number and returns a one-line verdict starting with
// "valid:" or "invalid:".
func validate(in string, strict bool) (string, bool) {
	d, err := mask.Normalize(in)
	if err != nil {
		return "invalid: " + describe(err), false
	}
	if !luhn.Valid(d) {
		return "invalid: luhn check failed", false
	}

	scheme, err := bin.Validate(d)
	switch {
	case err == nil:
		return fmt.Sprintf("valid: %s, %d digits", scheme, len(d)), true
	case errors.Is(err, bin.ErrUnknownScheme):
		if strict {
			return "invalid: unknown scheme", false
		}
		if len(d) < minPANLength || len(d) > maxPANLength {
			return fmt.Sprintf("invalid: unknown scheme and %d digits is not a card number length", len(d)), false
		}
		return fmt.Sprintf("valid: unknown scheme, %d digits", len(d)), true
	default:
		return "invalid: " + describe(err), false
	}
}
