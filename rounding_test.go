package geoskeleton

import (
	"fmt"
	"testing"
)

var PI = 3.141592653

// TestRoundToPrecisionSuccess
func TestRoundToPrecisionSuccess(t *testing.T) {
	// test
	num := RoundToPrecision(3.141592653, 2)
	str := fmt.Sprintf("%v", num)
	if 4 != len(str) {
		t.Error("Length of value is not what was expected")
	}
}

// BenchmarkRoundToPrecision
func BenchmarkRoundToPrecision(b *testing.B) {
	// benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RoundToPrecision(PI, 2)
	}
}

// TestRoundToPrecisionSuccess
func TestRoundSuccess(t *testing.T) {
	// test
	num := Round(PI)
	if 3 != num {
		t.Error("Unexpected value")
	}
}

// BenchmarkRoundToPrecision
func BenchmarkRound(b *testing.B) {
	// benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Round(PI)
	}
}
