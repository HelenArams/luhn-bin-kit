package bin

// Range is an inclusive range of leading digits assigned to a scheme. Lo and
// Hi always have the same number of digits, so a number falls inside the
// range when its first len(Lo) digits sort between them as strings.
type Range struct {
	Lo, Hi string
}

// Digits returns how many leading digits the range spans.
func (r Range) Digits() int {
	return len(r.Lo)
}

// Contains reports whether digits starts with a prefix inside the range.
// It expects an all-digit string; shorter inputs never match.
func (r Range) Contains(digits string) bool {
	n := len(r.Lo)
	if len(digits) < n {
		return false
	}
	p := digits[:n]
	return p >= r.Lo && p <= r.Hi
}

type entry struct {
	Range
	scheme Scheme
}

// table lists every known range, longest prefix first, so that the most
// specific assignment wins: Discover's 622126–622925 must be found before a
// shorter range that happens to start with 62. The order within one prefix
// length does not matter because ranges of equal length never overlap.
// TestTable enforces both properties.
var table = []entry{
	{Range{"622126", "622925"}, Discover},

	{Range{"2221", "2720"}, Mastercard},
	{Range{"3095", "3095"}, Diners},
	{Range{"3528", "3589"}, JCB},
	{Range{"6011", "6011"}, Discover},

	{Range{"300", "305"}, Diners},
	{Range{"644", "649"}, Discover},

	{Range{"34", "34"}, Amex},
	{Range{"36", "36"}, Diners},
	{Range{"37", "37"}, Amex},
	{Range{"38", "39"}, Diners},
	{Range{"51", "55"}, Mastercard},
	{Range{"65", "65"}, Discover},

	{Range{"4", "4"}, Visa},
}

// Ranges returns the prefix ranges assigned to s, most specific first. The
// slice is a copy and may be modified. Unknown has no ranges.
func Ranges(s Scheme) []Range {
	var out []Range
	for _, e := range table {
		if e.scheme == s {
			out = append(out, e.Range)
		}
	}
	return out
}
