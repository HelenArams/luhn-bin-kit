package mask

import (
	"errors"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"digits only", "4111111111111111", "4111111111111111"},
		{"spaces", "4111 1111 1111 1111", "4111111111111111"},
		{"hyphens", "4111-1111-1111-1111", "4111111111111111"},
		{"mixed separators", "4111 1111-1111 1111", "4111111111111111"},
		{"leading and trailing", " 4111111111111111 ", "4111111111111111"},
		{"doubled separators", "4111--1111  1111", "411111111111"},
		{"amex grouping", "3782 822463 10005", "378282246310005"},
		{"single digit", "4", "4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.in)
			if err != nil {
				t.Fatalf("Normalize(%q) error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		pos  int
		char rune
	}{
		{"letter", "4111x111", 4, 'x'},
		{"tab", "4111\t1111", 4, '\t'},
		{"dot", "4111.1111", 4, '.'},
		{"plus", "+4111", 0, '+'},
		{"fullwidth digit", "４111", 0, '４'},
		{"en dash", "4111–1111", 4, '–'},
		{"newline", "4111111111111111\n", 16, '\n'},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Normalize(tt.in)
			var ice *InvalidCharError
			if !errors.As(err, &ice) {
				t.Fatalf("Normalize(%q) error = %T (%v), want *InvalidCharError", tt.in, err, err)
			}
			if ice.Pos != tt.pos || ice.Char != tt.char {
				t.Fatalf("Normalize(%q) error = %+v, want Pos %d Char %q", tt.in, *ice, tt.pos, tt.char)
			}
		})
	}

	for _, in := range []string{"", " ", "-", " - - "} {
		if _, err := Normalize(in); !errors.Is(err, ErrEmpty) {
			t.Errorf("Normalize(%q) error = %v, want ErrEmpty", in, err)
		}
	}

	if got, want := (&InvalidCharError{Pos: 3, Char: 'x'}).Error(), `mask: invalid character 'x' at position 3`; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"visa 16", "4111111111111111", "411111******1111"},
		{"visa with spaces", "4111 1111 1111 1111", "411111******1111"},
		{"visa with hyphens", "4111-1111-1111-1111", "411111******1111"},
		{"amex 15", "378282246310005", "378282*****0005"},
		{"diners 14", "30569309025904", "305693****5904"},
		{"visa 13", "4222222222222", "422222***2222"},
		{"unionpay 19", "6205500000000000004", "620550*********0004"},
		{"eight digits", "12345678", "****5678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mask(tt.in)
			if err != nil {
				t.Fatalf("Mask(%q) error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("Mask(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaskErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{"empty", "", ErrEmpty},
		{"separators only", " - ", ErrEmpty},
		{"seven digits", "1234567", ErrTooShort},
		{"seven digits with spaces", "123 456 7", ErrTooShort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mask(tt.in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Mask(%q) error = %v, want %v", tt.in, err, tt.want)
			}
			if got != "" {
				t.Fatalf("Mask(%q) = %q on error, want empty", tt.in, got)
			}
		})
	}

	_, err := Mask("4111 1111 1111 111a")
	var ice *InvalidCharError
	if !errors.As(err, &ice) {
		t.Fatalf("Mask with letter error = %T (%v), want *InvalidCharError", err, err)
	}
}

func TestLast4(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"4111111111111111", "1111"},
		{"4111 1111 1111 1234", "1234"},
		{"378282246310005", "0005"},
		{"1234", "1234"},
	}
	for _, tt := range tests {
		got, err := Last4(tt.in)
		if err != nil {
			t.Errorf("Last4(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Last4(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}

	if _, err := Last4("123"); !errors.Is(err, ErrTooShort) {
		t.Errorf("Last4(\"123\") error = %v, want ErrTooShort", err)
	}
	if _, err := Last4(""); !errors.Is(err, ErrEmpty) {
		t.Errorf("Last4(\"\") error = %v, want ErrEmpty", err)
	}
	if _, err := Last4("12x4"); err == nil {
		t.Error("Last4 with a letter should fail")
	}
}
