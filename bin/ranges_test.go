package bin

import (
	"sort"
	"testing"
)

// TestTable pins the invariants Detect relies on: well-formed ranges,
// longest prefix first, and no overlap between ranges of the same length.
func TestTable(t *testing.T) {
	for i, e := range table {
		if e.Lo == "" || len(e.Lo) != len(e.Hi) {
			t.Errorf("table[%d] %v: Lo and Hi must have the same non-zero length", i, e.Range)
		}
		if !isDigits(e.Lo) || !isDigits(e.Hi) {
			t.Errorf("table[%d] %v: bounds must be digits", i, e.Range)
		}
		if e.Lo > e.Hi {
			t.Errorf("table[%d] %v: Lo must not exceed Hi", i, e.Range)
		}
		if e.scheme == Unknown {
			t.Errorf("table[%d] %v: assigned to Unknown", i, e.Range)
		}
		if i > 0 && e.Digits() > table[i-1].Digits() {
			t.Errorf("table[%d] %v: ranges must be ordered longest prefix first", i, e.Range)
		}
	}

	for i, a := range table {
		for _, b := range table[i+1:] {
			if a.Digits() != b.Digits() {
				continue
			}
			if a.Lo <= b.Hi && b.Lo <= a.Hi {
				t.Errorf("ranges %v (%v) and %v (%v) overlap", a.Range, a.scheme, b.Range, b.scheme)
			}
		}
	}
}

func TestEverySchemeIsFullyDescribed(t *testing.T) {
	for _, s := range Schemes() {
		if s == Unknown {
			t.Fatal("Schemes() must not include Unknown")
		}
		if len(Ranges(s)) == 0 {
			t.Errorf("%v has no ranges", s)
		}
		ls := s.Lengths()
		if len(ls) == 0 {
			t.Errorf("%v has no lengths", s)
		}
		if !sort.IntsAreSorted(ls) {
			t.Errorf("%v lengths %v are not ascending", s, ls)
		}
		for _, l := range ls {
			if l < 8 || l > 19 {
				t.Errorf("%v length %d is outside the ISO/IEC 7812 envelope", s, l)
			}
		}
		if s.String() == "" || s.ID() == "" {
			t.Errorf("%v is missing a name or id", uint8(s))
		}
	}
	if got, want := len(Schemes()), len(names)-1; got != want {
		t.Errorf("Schemes() lists %d schemes, names has %d", got, want)
	}
}

// TestCarveOuts checks the cases where a longer range belonging to one scheme
// sits inside a shorter range of another, which is exactly what the
// longest-prefix-first ordering exists for.
func TestCarveOuts(t *testing.T) {
	tests := []struct {
		in   string
		want Scheme
	}{
		{"62212600", Discover},
		{"62292599", Discover},
		{"62212599", UnionPay},
		{"62292600", UnionPay},
		{"62000000", UnionPay},
		{"60110000", Discover},
		{"63040000", Maestro},
		{"67590000", Maestro},
	}
	for _, tt := range tests {
		if got, _ := Detect(tt.in); got != tt.want {
			t.Errorf("Detect(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestRanges(t *testing.T) {
	got := Ranges(Discover)
	want := []Range{{"622126", "622925"}, {"6011", "6011"}, {"644", "649"}, {"65", "65"}}
	if len(got) != len(want) {
		t.Fatalf("Ranges(Discover) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Ranges(Discover)[%d] = %v, want %v", i, got[i], want[i])
		}
	}

	got[0] = Range{"0", "0"}
	if Ranges(Discover)[0] == got[0] {
		t.Fatal("Ranges must return a copy")
	}

	if r := Ranges(Unknown); r != nil {
		t.Fatalf("Ranges(Unknown) = %v, want nil", r)
	}
}

func TestRangeContains(t *testing.T) {
	r := Range{"300", "305"}
	tests := []struct {
		in   string
		want bool
	}{
		{"300", true},
		{"305", true},
		{"3029999", true},
		{"306", false},
		{"299", false},
		{"30", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := r.Contains(tt.in); got != tt.want {
			t.Errorf("%v.Contains(%q) = %v, want %v", r, tt.in, got, tt.want)
		}
	}
}
