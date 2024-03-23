package luhn

import "testing"

func FuzzValid(f *testing.F) {
	for _, s := range []string{"4111111111111111", "378282246310005", "", "0", "18", "4111 1111", "٤١١١"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		ok := Valid(s)
		if !ok || len(s) < 2 {
			if ok {
				t.Fatalf("Valid(%q) = true for input shorter than 2", s)
			}
			return
		}
		// Anything Valid accepts must agree with CheckDigit on its payload.
		got, err := CheckDigit(s[:len(s)-1])
		if err != nil {
			t.Fatalf("Valid(%q) = true but CheckDigit errored: %v", s, err)
		}
		if got != s[len(s)-1] {
			t.Fatalf("Valid(%q) = true but CheckDigit gives %q", s, got)
		}
	})
}
