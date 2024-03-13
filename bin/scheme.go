package bin

import (
	"errors"
	"fmt"
	"strings"
)

// Scheme identifies the card network that owns a number range.
type Scheme uint8

// Schemes known to the package. Unknown is the zero value and is what
// Detect reports when no range matches.
const (
	Unknown Scheme = iota
	Visa
	Mastercard
	Amex
	Discover
	JCB
	Diners
)

// names are the display names returned by String.
var names = [...]string{
	Unknown:    "Unknown",
	Visa:       "Visa",
	Mastercard: "Mastercard",
	Amex:       "American Express",
	Discover:   "Discover",
	JCB:        "JCB",
	Diners:     "Diners Club",
}

// ids are the stable lower-case identifiers returned by ID and used for
// text marshaling. Unlike display names they never contain spaces.
var ids = [...]string{
	Unknown:    "unknown",
	Visa:       "visa",
	Mastercard: "mastercard",
	Amex:       "amex",
	Discover:   "discover",
	JCB:        "jcb",
	Diners:     "diners",
}

// aliases are extra spellings accepted by ParseScheme, keyed by their
// normalized form (lower case, separators removed).
var aliases = map[string]Scheme{
	"mc":                      Mastercard,
	"master":                  Mastercard,
	"americanexpress":         Amex,
	"dinersclub":              Diners,
	"dinersclubinternational": Diners,
}

// lengths lists the number lengths each scheme issues, ascending.
var lengths = [...][]int{
	Visa:       {13, 16, 19},
	Mastercard: {16},
	Amex:       {15},
	Discover:   {16, 19},
	JCB:        {16, 17, 18, 19},
	Diners:     {14, 15, 16, 17, 18, 19},
}

// ErrUnknownScheme is returned by Validate when no range matches the number.
// Errors from ParseScheme and Scheme.UnmarshalText also match it with
// errors.Is.
var ErrUnknownScheme = errors.New("bin: unknown card scheme")

// UnknownSchemeError is returned by ParseScheme for a name that does not
// identify any scheme. It matches ErrUnknownScheme with errors.Is.
type UnknownSchemeError struct {
	Name string
}

func (e *UnknownSchemeError) Error() string {
	return fmt.Sprintf("bin: unknown card scheme %q", e.Name)
}

// Is reports whether target is ErrUnknownScheme, so callers can test for
// the condition without caring which function produced the error.
func (e *UnknownSchemeError) Is(target error) bool {
	return target == ErrUnknownScheme
}

// Schemes returns every scheme Detect can report, in a stable order that is
// convenient for help texts and tables. Unknown is not included.
func Schemes() []Scheme {
	return []Scheme{Visa, Mastercard, Amex, Discover, JCB, Diners}
}

// String returns the display name of the scheme, for example "American
// Express". Values outside the known set format as "Scheme(n)".
func (s Scheme) String() string {
	if int(s) < len(names) {
		return names[s]
	}
	return fmt.Sprintf("Scheme(%d)", uint8(s))
}

// ID returns the stable lower-case identifier of the scheme, for example
// "amex". It is the form ParseScheme always accepts and MarshalText emits.
// Values outside the known set return "".
func (s Scheme) ID() string {
	if int(s) < len(ids) {
		return ids[s]
	}
	return ""
}

// Lengths returns the number lengths the scheme issues, in ascending order.
// The slice is a copy and may be modified. Unknown has no lengths.
func (s Scheme) Lengths() []int {
	if int(s) >= len(lengths) {
		return nil
	}
	return append([]int(nil), lengths[s]...)
}

// ValidLength reports whether the scheme issues numbers of n digits.
func (s Scheme) ValidLength(n int) bool {
	if int(s) >= len(lengths) {
		return false
	}
	for _, l := range lengths[s] {
		if l == n {
			return true
		}
	}
	return false
}

// MarshalText implements encoding.TextMarshaler using the scheme's ID.
func (s Scheme) MarshalText() ([]byte, error) {
	id := s.ID()
	if id == "" {
		return nil, fmt.Errorf("bin: cannot marshal %s", s)
	}
	return []byte(id), nil
}

// UnmarshalText implements encoding.TextUnmarshaler by way of ParseScheme.
func (s *Scheme) UnmarshalText(text []byte) error {
	parsed, err := ParseScheme(string(text))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// ParseScheme resolves a scheme from its ID ("amex"), its display name
// ("American Express") or a common alias ("mc"). Matching ignores case,
// spaces, hyphens and underscores. An unrecognised name yields Unknown and
// an *UnknownSchemeError.
func ParseScheme(name string) (Scheme, error) {
	key := normalizeName(name)
	for s := Unknown; int(s) < len(ids); s++ {
		if ids[s] == key || normalizeName(names[s]) == key {
			return s, nil
		}
	}
	if s, ok := aliases[key]; ok {
		return s, nil
	}
	return Unknown, &UnknownSchemeError{Name: name}
}

func normalizeName(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range strings.ToLower(name) {
		switch r {
		case ' ', '-', '_':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
