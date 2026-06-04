package correlation

import (
	"math"
	"testing"
)

// Test against textbook values. Reference: Agresti, "An Introduction to Categorical
// Data Analysis", 2x2 examples.
func TestFishersExact(t *testing.T) {
	cases := []struct {
		a, b, c, d int
		want       float64
	}{
		{1, 9, 11, 3, 0.0027}, // strong association
		{3, 3, 3, 3, 1.0},     // symmetric, p must be 1
		{5, 0, 0, 5, 0.0079},  // perfect separation
	}
	for _, tc := range cases {
		got := fishersExact(tc.a, tc.b, tc.c, tc.d)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("fishersExact(%d,%d,%d,%d) = %.4f, want %.4f", tc.a, tc.b, tc.c, tc.d, got, tc.want)
		}
	}
}
