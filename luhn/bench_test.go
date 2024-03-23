package luhn

import (
	"strconv"
	"testing"
)

var sink bool

func BenchmarkValid(b *testing.B) {
	for _, pan := range []string{"4111111111111111", "6205500000000000004"} {
		b.Run(strconv.Itoa(len(pan)), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sink = Valid(pan)
			}
		})
	}
}
