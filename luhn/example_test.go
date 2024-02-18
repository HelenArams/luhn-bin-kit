package luhn_test

import (
	"fmt"
	"math/rand"

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

func ExampleGenerate() {
	// A fixed seed keeps the example reproducible; pass nil to use the
	// package-level math/rand source instead.
	rng := rand.New(rand.NewSource(2024))
	pan, err := luhn.Generate("411111", 16, rng)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(pan, luhn.Valid(pan))
	// Output:
	// 4111115088259407 true
}
