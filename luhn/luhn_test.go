package luhn

import (
	"errors"
	"testing"
)

func TestValid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"visa 16", "4111111111111111", true},
		{"visa 13", "4222222222222", true},
		{"mastercard", "5555555555554444", true},
		{"amex 15", "378282246310005", true},
		{"discover", "6011111111111117", true},
		{"jcb", "3530111333300000", true},
		{"diners 14", "30569309025904", true},
		{"imei", "490154203237518", true},
		{"two digits", "18", true},
		{"all zeros", "0000000000000000", true},
		{"last digit off by one", "4111111111111112", false},
		{"adjacent digits swapped", "4111111111111121", false},
		{"90 pair", "411111111111900", true},
		{"09 pair, the one transposition Luhn misses", "411111111111090", true},
		{"empty", "", false},
		{"single digit", "0", false},
		{"with spaces", "4111 1111 1111 1111", false},
		{"with dashes", "4111-1111-1111-1111", false},
		{"letters", "4111111111111111a", false},
		{"unicode digits", "４111111111111111", false},
		{"leading plus", "+4111111111111111", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Valid(tt.in); got != tt.want {
				t.Fatalf("Valid(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCheckDigit(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    byte
	}{
		{"visa", "411111111111111", '1'},
		{"mastercard", "555555555555444", '4'},
		{"amex", "37828224631000", '5'},
		{"discover", "601111111111111", '7'},
		{"single digit payload", "7", '5'},
		{"zero payload", "0", '0'},
		{"wraps to zero", "9", '1'},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckDigit(tt.payload)
			if err != nil {
				t.Fatalf("CheckDigit(%q) returned error: %v", tt.payload, err)
			}
			if got != tt.want {
				t.Fatalf("CheckDigit(%q) = %q, want %q", tt.payload, got, tt.want)
			}
			if full := tt.payload + string(got); !Valid(full) {
				t.Fatalf("payload + check digit %q does not pass Valid", full)
			}
		})
	}
}

func TestCheckDigitErrors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, err := CheckDigit("")
		if !errors.Is(err, ErrEmpty) {
			t.Fatalf("CheckDigit(\"\") error = %v, want ErrEmpty", err)
		}
	})

	t.Run("non-digit", func(t *testing.T) {
		_, err := CheckDigit("4111 1111")
		var nd *NonDigitError
		if !errors.As(err, &nd) {
			t.Fatalf("CheckDigit error = %T (%v), want *NonDigitError", err, err)
		}
		if nd.Pos != 4 || nd.Char != ' ' {
			t.Fatalf("NonDigitError = %+v, want Pos 4, Char ' '", *nd)
		}
		if got, want := nd.Error(), `luhn: non-digit character ' ' at position 4`; got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}

// TestCheckDigitRoundTrip checks the two halves of the package against each
// other: stripping the last digit of a valid number and recomputing it must
// give the digit back, and any other digit must make Valid fail.
func TestCheckDigitRoundTrip(t *testing.T) {
	numbers := []string{
		"4111111111111111",
		"4222222222222",
		"5555555555554444",
		"378282246310005",
		"6011111111111117",
		"3530111333300000",
		"30569309025904",
		"6759649826438453",
	}
	for _, n := range numbers {
		payload, check := n[:len(n)-1], n[len(n)-1]
		got, err := CheckDigit(payload)
		if err != nil {
			t.Fatalf("CheckDigit(%q): %v", payload, err)
		}
		if got != check {
			t.Errorf("CheckDigit(%q) = %q, want %q", payload, got, check)
		}
		for d := byte('0'); d <= '9'; d++ {
			if d == check {
				continue
			}
			if Valid(payload + string(d)) {
				t.Errorf("Valid(%q) = true, only %q should pass", payload+string(d), check)
			}
		}
	}
}
