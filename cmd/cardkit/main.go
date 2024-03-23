// Command cardkit validates, classifies and masks payment card numbers from
// the command line. It is a thin front end for the luhn, bin and mask
// packages of this module.
//
// Usage:
//
//	cardkit validate [-strict] [number...]
//	cardkit detect [-id] [number...]
//	cardkit mask [-last4] [number...]
//
// Numbers may contain spaces or hyphens. When no number is given on the
// command line they are read from standard input, one per line. When more
// than one number is processed each output line is prefixed with the
// masked input, so results can be told apart without echoing full numbers.
//
// The exit status is 0 when every number passed, 1 when at least one did
// not, and 2 for usage errors.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/HelenArams/luhn-bin-kit/mask"
)

const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

type command struct {
	name    string
	summary string
	run     func(args []string, stdin io.Reader, stdout, stderr io.Writer) int
}

var commands = []command{
	{"validate", "check the Luhn digit, scheme and length of numbers", cmdValidate},
	{"detect", "print the card scheme that owns each number or BIN", cmdDetect},
	{"mask", "hide all but the first six and last four digits", cmdMask},
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return exitUsage
	}
	switch args[0] {
	case "help", "-h", "-help", "--help":
		usage(stdout)
		return exitOK
	}
	for _, c := range commands {
		if c.name == args[0] {
			return c.run(args[1:], stdin, stdout, stderr)
		}
	}
	fmt.Fprintf(stderr, "cardkit: unknown command %q\n\n", args[0])
	usage(stderr)
	return exitUsage
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "Usage: cardkit <command> [flags] [number...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, c := range commands {
		fmt.Fprintf(w, "  %-9s %s\n", c.name, c.summary)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Numbers may contain spaces or hyphens. Without arguments, numbers are read")
	fmt.Fprintln(w, "from standard input, one per line. Run 'cardkit <command> -h' for flags.")
}

func newFlagSet(name, synopsis string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: cardkit %s %s\n", name, synopsis)
		fs.PrintDefaults()
	}
	return fs
}

// parseFlags reports whether the command should go on. If not, code is the
// exit status: 0 after -h, 2 after a bad flag (flag already printed usage).
func parseFlags(fs *flag.FlagSet, args []string) (code int, ok bool) {
	err := fs.Parse(args)
	switch {
	case err == nil:
		return exitOK, true
	case errors.Is(err, flag.ErrHelp):
		return exitOK, false
	default:
		return exitUsage, false
	}
}

// each runs fn over the numbers in args, or over stdin lines when there are
// none. With labelled set every line gets the masked input as a prefix so
// you can tell results apart without the tool echoing full PANs.
func each(args []string, stdin io.Reader, stdout, stderr io.Writer, labelled bool, fn func(in string) (out string, ok bool)) int {
	code := exitOK
	handle := func(in string) {
		out, ok := fn(in)
		if !ok {
			code = exitFail
		}
		if labelled {
			fmt.Fprintf(stdout, "%s\t%s\n", label(in), out)
		} else {
			fmt.Fprintln(stdout, out)
		}
	}

	if len(args) > 0 {
		for _, in := range args {
			handle(in)
		}
		return code
	}

	sc := bufio.NewScanner(stdin)
	for sc.Scan() {
		in := strings.TrimSpace(sc.Text())
		if in == "" {
			continue
		}
		handle(in)
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(stderr, "cardkit: reading standard input: %v\n", err)
		return exitUsage
	}
	return code
}

// label is the masked input, or the raw input when it cannot be masked (a
// bare BIN, garbage) - in which case it was never a full PAN anyway.
func label(in string) string {
	if m, err := mask.Mask(in); err == nil {
		return m
	}
	return in
}

// describe drops the "pkg: " prefix from library errors; it's useful for Go
// callers but just noise on a terminal.
func describe(err error) string {
	msg := err.Error()
	if i := strings.Index(msg, ": "); i > 0 && !strings.ContainsAny(msg[:i], " \"") {
		msg = msg[i+2:]
	}
	return msg
}
