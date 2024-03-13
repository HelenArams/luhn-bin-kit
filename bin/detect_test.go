package bin

import (
	"errors"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Scheme
		ok   bool
	}{
		// Full numbers.
		{"visa 16", "4111111111111111", Visa, true},
		{"visa 13", "4222222222222", Visa, true},
		{"visa 19", "4917610000000000003", Visa, true},
		{"mastercard 5-series", "5555555555554444", Mastercard, true},
		{"mastercard 2-series low", "2221000000000009", Mastercard, true},
		{"mastercard 2-series", "2223000048400011", Mastercard, true},
		{"amex 34", "343434343434343", Amex, true},
		{"amex 37", "378282246310005", Amex, true},
		{"discover 6011", "6011111111111117", Discover, true},
		{"discover 65", "6500000000000002", Discover, true},
		{"discover 644-649", "6445644564456445", Discover, true},
		{"discover 622126", "6221260000000000", Discover, true},
		{"discover 622925 upper bound", "6229250000000000", Discover, true},
		{"jcb", "3530111333300000", JCB, true},
		{"jcb upper", "3589000000000000", JCB, true},
		{"diners 36", "36700102000000", Diners, true},
		{"diners 38", "38520000023237", Diners, true},
		{"diners 30x", "30569309025904", Diners, true},
		{"diners 3095", "3095000000000000", Diners, true},
		{"unionpay 62", "6200000000000005", UnionPay, true},
		{"unionpay 19", "6205500000000000004", UnionPay, true},
		{"unionpay just below discover carve-out", "6221250000000000", UnionPay, true},
		{"unionpay just above discover carve-out", "6229260000000000", UnionPay, true},
		{"maestro 50", "5018000000000009", Maestro, true},
		{"maestro 56", "5600000000000000", Maestro, true},
		{"maestro 58", "5893000000000000", Maestro, true},
		{"maestro 6304", "6304000000000000", Maestro, true},
		{"maestro 67", "6759649826438453", Maestro, true},
		{"maestro 19", "6799990100000000019", Maestro, true},
		{"maestro 12", "501800000000", Maestro, true},

		// Prefixes only (BIN lookups).
		{"bin visa", "4", Visa, true},
		{"bin mastercard", "51", Mastercard, true},
		{"bin amex", "37", Amex, true},
		{"bin jcb 6 digits", "352800", JCB, true},
		{"bin discover 8 digits", "60110000", Discover, true},

		// Boundaries that must not match a neighbouring range.
		{"2220 below mastercard", "2220000000000000", Unknown, false},
		{"2721 above mastercard", "2721000000000000", Unknown, false},
		{"3527 below jcb", "3527000000000000", Unknown, false},
		{"3590 above jcb", "3590000000000000", Unknown, false},
		{"306 outside diners", "3060000000000000", Unknown, false},
		{"3094 not diners", "3094000000000000", Unknown, false},
		{"6010 not discover", "6010000000000000", Unknown, false},
		{"643 not discover", "6430000000000000", Unknown, false},
		{"59 unassigned", "5900000000000000", Unknown, false},
		{"60 other than 6011 unassigned", "6000000000000000", Unknown, false},
		{"63 other than 6304 unassigned", "6300000000000000", Unknown, false},
		{"66 unassigned", "6600000000000000", Unknown, false},
		{"35 alone is too short to decide", "35", Unknown, false},

		// Garbage.
		{"empty", "", Unknown, false},
		{"spaces", "4111 1111 1111 1111", Unknown, false},
		{"dashes", "4111-1111-1111-1111", Unknown, false},
		{"letters", "4a11111111111111", Unknown, false},
		{"unknown prefix", "9999999999999999", Unknown, false},
		{"one", "1", Unknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Detect(tt.in)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("Detect(%q) = %v, %v; want %v, %v", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Scheme
		wantErr error
		msg     string
	}{
		{"visa 16", "4111111111111111", Visa, nil, ""},
		{"visa 13", "4222222222222", Visa, nil, ""},
		{"amex 15", "378282246310005", Amex, nil, ""},
		{"diners 14", "30569309025904", Diners, nil, ""},
		{"jcb 19", "3530111333300000000", JCB, nil, ""},
		{"visa 15 digits", "411111111111111", Visa, &LengthError{Visa, 15},
			"bin: Visa numbers have 13, 16 or 19 digits, got 15"},
		{"amex 16 digits", "3782822463100050", Amex, &LengthError{Amex, 16},
			"bin: American Express numbers have 15 digits, got 16"},
		{"discover 17 digits", "60111111111111170", Discover, &LengthError{Discover, 17},
			"bin: Discover numbers have 16 or 19 digits, got 17"},
		{"diners 13 digits", "3056930902590", Diners, &LengthError{Diners, 13},
			"bin: Diners Club numbers have 14 to 19 digits, got 13"},
		{"maestro 12 digits", "501800000000", Maestro, nil, ""},
		{"maestro 11 digits", "50180000000", Maestro, &LengthError{Maestro, 11},
			"bin: Maestro numbers have 12 to 19 digits, got 11"},
		{"unionpay 15 digits", "620000000000000", UnionPay, &LengthError{UnionPay, 15},
			"bin: UnionPay numbers have 16 to 19 digits, got 15"},
		{"mastercard bin only", "5555", Mastercard, &LengthError{Mastercard, 4},
			"bin: Mastercard numbers have 16 digits, got 4"},
		{"unknown", "9999999999999999", Unknown, ErrUnknownScheme, "bin: unknown card scheme"},
		{"empty", "", Unknown, ErrNonDigit, "bin: number must contain only digits"},
		{"spaces", "4111 1111 1111 1111", Unknown, ErrNonDigit, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Validate(tt.in)
			if got != tt.want {
				t.Fatalf("Validate(%q) scheme = %v, want %v", tt.in, got, tt.want)
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate(%q) error = %v, want nil", tt.in, err)
				}
				return
			}
			var le *LengthError
			if errors.As(tt.wantErr, &le) {
				var gotLE *LengthError
				if !errors.As(err, &gotLE) || *gotLE != *le {
					t.Fatalf("Validate(%q) error = %#v, want %#v", tt.in, err, le)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate(%q) error = %v, want %v", tt.in, err, tt.wantErr)
			}
			if tt.msg != "" && err.Error() != tt.msg {
				t.Fatalf("Validate(%q) error text = %q, want %q", tt.in, err.Error(), tt.msg)
			}
		})
	}
}

func TestDescribeLengths(t *testing.T) {
	tests := []struct {
		in   []int
		want string
	}{
		{nil, "no"},
		{[]int{15}, "15"},
		{[]int{16, 19}, "16 or 19"},
		{[]int{16, 17}, "16 or 17"},
		{[]int{13, 16, 19}, "13, 16 or 19"},
		{[]int{16, 17, 18, 19}, "16 to 19"},
		{[]int{12, 13, 15}, "12, 13 or 15"},
	}
	for _, tt := range tests {
		if got := describeLengths(tt.in); got != tt.want {
			t.Errorf("describeLengths(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
