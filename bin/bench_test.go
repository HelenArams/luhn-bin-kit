package bin

import "testing"

var sinkScheme Scheme

func BenchmarkDetect(b *testing.B) {
	cases := map[string]string{
		"visa":     "4111111111111111", // last entry in the table, worst case
		"discover": "6221260000000000", // first entry
		"unknown":  "9999999999999999",
	}
	for name, pan := range cases {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sinkScheme, _ = Detect(pan)
			}
		})
	}
}
