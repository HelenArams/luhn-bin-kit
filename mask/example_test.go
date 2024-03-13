package mask_test

import (
	"errors"
	"fmt"

	"github.com/HelenArams/luhn-bin-kit/mask"
)

func ExampleMask() {
	masked, err := mask.Mask("4111 1111 1111 1111")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(masked)
	// Output:
	// 411111******1111
}

func ExampleLast4() {
	last4, _ := mask.Last4("3782-822463-10005")
	fmt.Println(last4)
	// Output:
	// 0005
}

func ExampleNormalize() {
	digits, err := mask.Normalize("4111-1111 1111 1111")
	fmt.Println(digits, err)

	_, err = mask.Normalize("4111.1111")
	var ice *mask.InvalidCharError
	if errors.As(err, &ice) {
		fmt.Printf("%q at %d\n", ice.Char, ice.Pos)
	}
	// Output:
	// 4111111111111111 <nil>
	// '.' at 4
}
