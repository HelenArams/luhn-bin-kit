package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type result struct {
	code   int
	stdout string
	stderr string
}

func exec(t *testing.T, stdin string, args ...string) result {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(args, strings.NewReader(stdin), &out, &errb)
	return result{code, out.String(), errb.String()}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		stdin      string
		code       int
		stdout     string // exact match when non-empty
		stderrHas  string
		stdoutHas  string
		stderrNone bool
	}{
		{
			name: "no arguments prints usage", args: nil,
			code: exitUsage, stderrHas: "Usage: cardkit <command>",
		},
		{
			name: "help", args: []string{"help"},
			code: exitOK, stdoutHas: "validate", stderrNone: true,
		},
		{
			name: "dash h", args: []string{"-h"},
			code: exitOK, stdoutHas: "Commands:",
		},
		{
			name: "unknown command", args: []string{"frobnicate"},
			code: exitUsage, stderrHas: `unknown command "frobnicate"`,
		},
		{
			name: "subcommand help", args: []string{"validate", "-h"},
			code: exitOK, stderrHas: "Usage: cardkit validate [-strict] [number...]",
		},
		{
			name: "subcommand bad flag", args: []string{"detect", "-bogus"},
			code: exitUsage, stderrHas: "flag provided but not defined: -bogus",
		},

		// validate
		{
			name: "validate visa", args: []string{"validate", "4111111111111111"},
			code: exitOK, stdout: "valid: Visa, 16 digits\n", stderrNone: true,
		},
		{
			name: "validate with separators", args: []string{"validate", "3782-822463-10005"},
			code: exitOK, stdout: "valid: American Express, 15 digits\n",
		},
		{
			name: "validate bad luhn", args: []string{"validate", "4111111111111112"},
			code: exitFail, stdout: "invalid: luhn check failed\n",
		},
		{
			name: "validate wrong length", args: []string{"validate", "411111111111116"},
			code: exitFail, stdout: "invalid: Visa numbers have 13, 16 or 19 digits, got 15\n",
		},
		{
			name: "validate letters", args: []string{"validate", "4111 1111 1111 111x"},
			code: exitFail, stdout: "invalid: invalid character 'x' at position 18\n",
		},
		{
			name: "validate unknown scheme", args: []string{"validate", "9999999999999995"},
			code: exitOK, stdout: "valid: unknown scheme, 16 digits\n",
		},
		{
			name: "validate unknown scheme strict", args: []string{"validate", "-strict", "9999999999999995"},
			code: exitFail, stdout: "invalid: unknown scheme\n",
		},
		{
			name: "validate unknown scheme implausible length", args: []string{"validate", "18"},
			code: exitFail, stdout: "invalid: unknown scheme and 2 digits is not a card number length\n",
		},
		{
			name: "validate several numbers are labelled", args: []string{"validate", "4111111111111111", "5555555555554445"},
			code: exitFail, stdout: "411111******1111\tvalid: Visa, 16 digits\n555555******4445\tinvalid: luhn check failed\n",
		},
		{
			name: "validate from stdin", args: []string{"validate"},
			stdin: "4111 1111 1111 1111\n\n  6011111111111117 \n",
			code:  exitOK, stdout: "411111******1111\tvalid: Visa, 16 digits\n601111******1117\tvalid: Discover, 16 digits\n",
		},
		{
			name: "validate garbage keeps raw label", args: []string{"validate", "abc", "4111111111111111"},
			code: exitFail, stdout: "abc\tinvalid: invalid character 'a' at position 0\n411111******1111\tvalid: Visa, 16 digits\n",
		},

		// detect
		{
			name: "detect visa", args: []string{"detect", "4111111111111111"},
			code: exitOK, stdout: "Visa\n",
		},
		{
			name: "detect bin only", args: []string{"detect", "3782"},
			code: exitOK, stdout: "American Express\n",
		},
		{
			name: "detect id", args: []string{"detect", "-id", "3782"},
			code: exitOK, stdout: "amex\n",
		},
		{
			name: "detect unknown", args: []string{"detect", "9999"},
			code: exitFail, stdout: "unknown\n",
		},
		{
			name: "detect invalid input", args: []string{"detect", "4111x"},
			code: exitFail, stdout: "error: invalid character 'x' at position 4\n",
		},
		{
			name: "detect several", args: []string{"detect", "5555", "6759649826438453"},
			code: exitOK, stdout: "5555\tMastercard\n675964******8453\tMaestro\n",
		},

		// mask
		{
			name: "mask", args: []string{"mask", "4111 1111 1111 1111"},
			code: exitOK, stdout: "411111******1111\n",
		},
		{
			name: "mask last4", args: []string{"mask", "-last4", "4111 1111 1111 1234"},
			code: exitOK, stdout: "1234\n",
		},
		{
			name: "mask too short", args: []string{"mask", "1234"},
			code: exitFail, stdout: "error: number too short\n",
		},
		{
			name: "mask several are not labelled", args: []string{"mask", "4111111111111111", "378282246310005"},
			code: exitOK, stdout: "411111******1111\n378282*****0005\n",
		},
		{
			name: "mask from stdin", args: []string{"mask"}, stdin: "4111111111111111\n378282246310005\n",
			code: exitOK, stdout: "411111******1111\n378282*****0005\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exec(t, tt.stdin, tt.args...)
			if got.code != tt.code {
				t.Errorf("exit code = %d, want %d (stdout %q, stderr %q)", got.code, tt.code, got.stdout, got.stderr)
			}
			if tt.stdout != "" && got.stdout != tt.stdout {
				t.Errorf("stdout = %q, want %q", got.stdout, tt.stdout)
			}
			if tt.stdoutHas != "" && !strings.Contains(got.stdout, tt.stdoutHas) {
				t.Errorf("stdout = %q, want it to contain %q", got.stdout, tt.stdoutHas)
			}
			if tt.stderrHas != "" && !strings.Contains(got.stderr, tt.stderrHas) {
				t.Errorf("stderr = %q, want it to contain %q", got.stderr, tt.stderrHas)
			}
			if tt.stderrNone && got.stderr != "" {
				t.Errorf("stderr = %q, want empty", got.stderr)
			}
		})
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("disk on fire") }

func TestStdinReadError(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"validate"}, failingReader{}, &out, &errb)
	if code != exitUsage {
		t.Fatalf("exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(errb.String(), "disk on fire") {
		t.Fatalf("stderr = %q, want the read error", errb.String())
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"bin: unknown card scheme", "unknown card scheme"},
		{"mask: invalid character 'x' at position 3", "invalid character 'x' at position 3"},
		{"no prefix here", "no prefix here"},
		{`bin: unknown card scheme "a: b"`, `unknown card scheme "a: b"`},
		{`odd "quoted: thing" first`, `odd "quoted: thing" first`},
	}
	for _, tt := range tests {
		if got := describe(errors.New(tt.in)); got != tt.want {
			t.Errorf("describe(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

var _ io.Reader = failingReader{}
