package bin_test

import (
	"errors"
	"fmt"

	"github.com/HelenArams/luhn-bin-kit/bin"
)

func ExampleDetect() {
	for _, pan := range []string{"4111111111111111", "378282246310005", "6200000000000005", "6221260000000000", "9999"} {
		scheme, ok := bin.Detect(pan)
		fmt.Println(scheme, ok)
	}
	// Output:
	// Visa true
	// American Express true
	// UnionPay true
	// Discover true
	// Unknown false
}

func ExampleDetect_bin() {
	// A BIN alone is enough; the rest of the number is not needed.
	scheme, _ := bin.Detect("5555")
	fmt.Println(scheme)
	// Output:
	// Mastercard
}

func ExampleValidate() {
	_, err := bin.Validate("411111111111111") // 15 digits
	var le *bin.LengthError
	if errors.As(err, &le) {
		fmt.Println(le.Scheme, "lengths:", le.Scheme.Lengths())
	}
	fmt.Println(err)

	_, err = bin.Validate("9999999999999999")
	fmt.Println(errors.Is(err, bin.ErrUnknownScheme))
	// Output:
	// Visa lengths: [13 16 19]
	// bin: Visa numbers have 13, 16 or 19 digits, got 15
	// true
}

func ExampleParseScheme() {
	s, err := bin.ParseScheme("American Express")
	fmt.Println(s, s.ID(), err)

	_, err = bin.ParseScheme("laser")
	fmt.Println(err)
	// Output:
	// American Express amex <nil>
	// bin: unknown card scheme "laser"
}

func ExampleScheme_Lengths() {
	fmt.Println(bin.Visa.Lengths(), bin.Amex.ValidLength(15), bin.Amex.ValidLength(16))
	// Output:
	// [13 16 19] true false
}
