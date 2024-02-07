package luhn_test

import (
	"fmt"

	"github.com/HelenArams/luhn-bin-kit/luhn"
)

func ExampleValid() {
	fmt.Println(luhn.Valid("4111111111111111"))
	fmt.Println(luhn.Valid("4111111111111112"))
	fmt.Println(luhn.Valid("4111 1111 1111 1111")) // separators are not stripped
	// Output:
	// true
	// false
	// false
}

func ExampleCheckDigit() {
	digit, err := luhn.CheckDigit("411111111111111")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%c\n", digit)
	// Output:
	// 1
}
