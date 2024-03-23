package mask

import (
	"strings"
	"testing"
)

func FuzzMask(f *testing.F) {
	for _, s := range []string{"4111 1111 1111 1111", "4111-1111-1111-1111", "12345678", "1234567", "", "--", "4111\t1111", "４１１１"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got, err := Mask(s)
		if err != nil {
			return
		}
		d, err := Normalize(s)
		if err != nil {
			t.Fatalf("Mask(%q) ok but Normalize failed: %v", s, err)
		}
		if len(got) != len(d) {
			t.Fatalf("Mask(%q) = %q, length differs from %q", s, got, d)
		}
		// The important property: never reveal more than 6+4 digits.
		if shown := len(got) - strings.Count(got, "*"); shown > 10 {
			t.Fatalf("Mask(%q) = %q reveals %d digits", s, got, shown)
		}
		if got[len(got)-4:] != d[len(d)-4:] {
			t.Fatalf("Mask(%q) = %q lost the last four", s, got)
		}
	})
}
