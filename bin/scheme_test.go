package bin

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestSchemeStringAndID(t *testing.T) {
	tests := []struct {
		s    Scheme
		name string
		id   string
	}{
		{Unknown, "Unknown", "unknown"},
		{Visa, "Visa", "visa"},
		{Mastercard, "Mastercard", "mastercard"},
		{Amex, "American Express", "amex"},
		{Discover, "Discover", "discover"},
		{JCB, "JCB", "jcb"},
		{Diners, "Diners Club", "diners"},
		{Scheme(200), "Scheme(200)", ""},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.name {
			t.Errorf("Scheme(%d).String() = %q, want %q", uint8(tt.s), got, tt.name)
		}
		if got := tt.s.ID(); got != tt.id {
			t.Errorf("Scheme(%d).ID() = %q, want %q", uint8(tt.s), got, tt.id)
		}
	}
}

func TestParseScheme(t *testing.T) {
	tests := []struct {
		in   string
		want Scheme
	}{
		{"visa", Visa},
		{"VISA", Visa},
		{"Visa", Visa},
		{"mastercard", Mastercard},
		{"MasterCard", Mastercard},
		{"master-card", Mastercard},
		{"mc", Mastercard},
		{"amex", Amex},
		{"American Express", Amex},
		{"american_express", Amex},
		{"discover", Discover},
		{"jcb", JCB},
		{"diners", Diners},
		{"Diners Club", Diners},
		{"diners club international", Diners},
		{"unknown", Unknown},
	}
	for _, tt := range tests {
		got, err := ParseScheme(tt.in)
		if err != nil {
			t.Errorf("ParseScheme(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseScheme(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}

	for _, s := range Schemes() {
		if got, err := ParseScheme(s.ID()); err != nil || got != s {
			t.Errorf("ParseScheme(%v.ID()) = %v, %v", s, got, err)
		}
		if got, err := ParseScheme(s.String()); err != nil || got != s {
			t.Errorf("ParseScheme(%v.String()) = %v, %v", s, got, err)
		}
	}
}

func TestParseSchemeUnknown(t *testing.T) {
	for _, in := range []string{"", "bitcoin", "vis", "4111111111111111"} {
		got, err := ParseScheme(in)
		if got != Unknown {
			t.Errorf("ParseScheme(%q) = %v, want Unknown", in, got)
		}
		var use *UnknownSchemeError
		if !errors.As(err, &use) || use.Name != in {
			t.Errorf("ParseScheme(%q) error = %#v, want *UnknownSchemeError with Name %q", in, err, in)
		}
		if !errors.Is(err, ErrUnknownScheme) {
			t.Errorf("ParseScheme(%q) error does not match ErrUnknownScheme", in)
		}
	}
	if got, want := (&UnknownSchemeError{Name: "foo"}).Error(), `bin: unknown card scheme "foo"`; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestSchemeTextRoundTrip(t *testing.T) {
	type doc struct {
		Scheme Scheme `json:"scheme"`
	}
	for _, s := range append(Schemes(), Unknown) {
		data, err := json.Marshal(doc{s})
		if err != nil {
			t.Fatalf("marshal %v: %v", s, err)
		}
		if want := `{"scheme":"` + s.ID() + `"}`; string(data) != want {
			t.Errorf("marshal %v = %s, want %s", s, data, want)
		}
		var back doc
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatalf("unmarshal %s: %v", data, err)
		}
		if back.Scheme != s {
			t.Errorf("round trip of %v gave %v", s, back.Scheme)
		}
	}

	var d doc
	err := json.Unmarshal([]byte(`{"scheme":"laser"}`), &d)
	if !errors.Is(err, ErrUnknownScheme) {
		t.Errorf("unmarshal of unknown scheme error = %v, want ErrUnknownScheme", err)
	}

	if _, err := Scheme(99).MarshalText(); err == nil {
		t.Error("marshal of out-of-range scheme should fail")
	}
}

func TestLengths(t *testing.T) {
	ls := Visa.Lengths()
	if len(ls) != 3 || ls[0] != 13 || ls[1] != 16 || ls[2] != 19 {
		t.Fatalf("Visa.Lengths() = %v", ls)
	}
	ls[0] = 1
	if Visa.Lengths()[0] != 13 {
		t.Fatal("Lengths must return a copy")
	}
	if Unknown.Lengths() != nil || Scheme(99).Lengths() != nil {
		t.Fatal("Unknown and out-of-range schemes must have no lengths")
	}

	tests := []struct {
		s    Scheme
		n    int
		want bool
	}{
		{Visa, 16, true},
		{Visa, 15, false},
		{Amex, 15, true},
		{Amex, 16, false},
		{Diners, 14, true},
		{Diners, 19, true},
		{Diners, 13, false},
		{Unknown, 16, false},
		{Scheme(99), 16, false},
	}
	for _, tt := range tests {
		if got := tt.s.ValidLength(tt.n); got != tt.want {
			t.Errorf("%v.ValidLength(%d) = %v, want %v", tt.s, tt.n, got, tt.want)
		}
	}
}
