package luhn

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

// fixedSource always returns the same digit, which makes the middle of a
// generated number predictable.
type fixedSource int

func (f fixedSource) Intn(int) int { return int(f) }

func TestGenerate(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		length int
		src    Source
		want   string // empty when only shape is checked
	}{
		{"visa 16 all sevens", "4", 16, fixedSource(7), "4777777777777778"},
		{"amex 15 zeros", "37", 15, fixedSource(0), "370000000000002"},
		{"prefix plus check digit only", "41111111111111", 15, fixedSource(9), "411111111111116"},
		{"empty prefix", "", 10, fixedSource(3), "3333333338"},
		{"seeded rand", "5555", 16, rand.New(rand.NewSource(1)), ""},
		{"nil source", "6011", 16, nil, ""},
		{"long number", "62", 19, nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.prefix, tt.length, tt.src)
			if err != nil {
				t.Fatalf("Generate(%q, %d) returned error: %v", tt.prefix, tt.length, err)
			}
			if len(got) != tt.length {
				t.Fatalf("Generate(%q, %d) = %q, want %d digits", tt.prefix, tt.length, got, tt.length)
			}
			if !strings.HasPrefix(got, tt.prefix) {
				t.Fatalf("Generate(%q, %d) = %q, prefix lost", tt.prefix, tt.length, got)
			}
			if !Valid(got) {
				t.Fatalf("Generate(%q, %d) = %q, fails Valid", tt.prefix, tt.length, got)
			}
			if tt.want != "" && got != tt.want {
				t.Fatalf("Generate(%q, %d) = %q, want %q", tt.prefix, tt.length, got, tt.want)
			}
		})
	}
}

func TestGenerateIsDeterministicForSeed(t *testing.T) {
	a, err := Generate("4", 16, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate("4", 16, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("same seed produced %q and %q", a, b)
	}
}

func TestGenerateErrors(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		length int
		want   error
	}{
		{"length equals prefix", "4111", 4, ErrLength},
		{"length shorter than prefix", "4111", 2, ErrLength},
		{"zero length", "", 0, ErrLength},
		{"negative length", "4", -1, ErrLength},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.prefix, tt.length, nil)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Generate(%q, %d) error = %v, want %v", tt.prefix, tt.length, err, tt.want)
			}
			if got != "" {
				t.Fatalf("Generate(%q, %d) = %q on error, want empty", tt.prefix, tt.length, got)
			}
		})
	}

	t.Run("non-digit prefix", func(t *testing.T) {
		_, err := Generate("4x11", 16, nil)
		var nd *NonDigitError
		if !errors.As(err, &nd) {
			t.Fatalf("error = %T (%v), want *NonDigitError", err, err)
		}
		if nd.Pos != 1 || nd.Char != 'x' {
			t.Fatalf("NonDigitError = %+v, want Pos 1, Char 'x'", *nd)
		}
	})
}
