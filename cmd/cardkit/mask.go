package main

import (
	"io"

	"github.com/HelenArams/luhn-bin-kit/mask"
)

func cmdMask(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("mask", "[-last4] [number...]", stderr)
	last4 := fs.Bool("last4", false, "print only the last four digits")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	// The masked output identifies the input well enough on its own, so
	// mask never adds a label.
	return each(fs.Args(), stdin, stdout, stderr, false, func(in string) (string, bool) {
		var out string
		var err error
		if *last4 {
			out, err = mask.Last4(in)
		} else {
			out, err = mask.Mask(in)
		}
		if err != nil {
			return "error: " + describe(err), false
		}
		return out, true
	})
}
