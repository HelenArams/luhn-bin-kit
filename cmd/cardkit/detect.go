package main

import (
	"io"

	"github.com/HelenArams/luhn-bin-kit/bin"
	"github.com/HelenArams/luhn-bin-kit/mask"
)

func cmdDetect(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("detect", "[-id] [number...]", stderr)
	id := fs.Bool("id", false, "print the scheme identifier (e.g. amex) instead of its name")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	return each(fs.Args(), stdin, stdout, stderr, len(fs.Args()) != 1, func(in string) (string, bool) {
		return detect(in, *id)
	})
}

func detect(in string, id bool) (string, bool) {
	d, err := mask.Normalize(in)
	if err != nil {
		return "error: " + describe(err), false
	}
	s, ok := bin.Detect(d)
	if !ok {
		return "unknown", false
	}
	if id {
		return s.ID(), true
	}
	return s.String(), true
}
