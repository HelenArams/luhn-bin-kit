// Package luhn implements the Luhn (mod 10) checksum used by payment card
// numbers, IMEIs and a number of other identifiers.
//
// The package works on strings of ASCII digits and deliberately does not
// strip separators: callers that accept user input should normalize it
// first (for example with the mask package's Normalize) so that validation
// failures are explicit rather than silently tolerated.
//
// Valid checks a complete number, CheckDigit computes the digit that must be
// appended to a payload, and Generate produces synthetic numbers for test
// fixtures. None of the functions allocate on the happy path except
// Generate, which returns a new string.
package luhn
