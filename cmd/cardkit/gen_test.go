package main

import (
	"strings"
	"testing"

	"github.com/HelenArams/luhn-bin-kit/bin"
	"github.com/HelenArams/luhn-bin-kit/luhn"
)

func TestGen(t *testing.T) {
	for _, s := range bin.Schemes() {
		t.Run(s.ID(), func(t *testing.T) {
			got := exec(t, "", "gen", "--scheme", s.ID(), "--count", "50", "--seed", "7")
			if got.code != exitOK {
				t.Fatalf("exit %d, stderr %q", got.code, got.stderr)
			}
			lines := strings.Split(strings.TrimSpace(got.stdout), "\n")
			if len(lines) != 50 {
				t.Fatalf("got %d lines, want 50", len(lines))
			}
			for _, pan := range lines {
				if !luhn.Valid(pan) {
					t.Errorf("%s fails luhn", pan)
				}
				if sc, err := bin.Validate(pan); err != nil || sc != s {
					t.Errorf("%s detected as %v (err %v), want %v", pan, sc, err, s)
				}
			}
		})
	}
}

func TestGenSeedIsReproducible(t *testing.T) {
	a := exec(t, "", "gen", "-scheme", "mastercard", "-count", "3", "-seed", "42")
	b := exec(t, "", "gen", "-scheme", "mastercard", "-count", "3", "-seed", "42")
	if a.stdout != b.stdout || a.stdout == "" {
		t.Fatalf("same seed gave %q and %q", a.stdout, b.stdout)
	}
}

func TestGenLength(t *testing.T) {
	got := exec(t, "", "gen", "-scheme", "visa", "-length", "19", "-seed", "1")
	if pan := strings.TrimSpace(got.stdout); len(pan) != 19 {
		t.Fatalf("got %q, want 19 digits", pan)
	}
	got = exec(t, "", "gen", "-scheme", "amex", "-seed", "1")
	if pan := strings.TrimSpace(got.stdout); len(pan) != 15 {
		t.Fatalf("amex default length: got %q", pan)
	}
}

func TestGenErrors(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"gen", "-scheme", "laser"}, `unknown scheme "laser"`},
		{[]string{"gen", "-scheme", "unknown"}, `unknown scheme "unknown"`},
		{[]string{"gen", "-count", "0"}, "-count must be between"},
		{[]string{"gen", "-scheme", "amex", "-length", "16"}, "can't be 16 digits"},
		{[]string{"gen", "4111"}, `unexpected argument "4111"`},
	}
	for _, tt := range tests {
		got := exec(t, "", tt.args...)
		if got.code != exitUsage || !strings.Contains(got.stderr, tt.want) {
			t.Errorf("%v: exit %d stderr %q, want usage error containing %q", tt.args, got.code, got.stderr, tt.want)
		}
	}
}
