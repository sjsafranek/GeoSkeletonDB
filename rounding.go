package geoskeleton

import "math"

// Round float64.
// Source: https://gist.github.com/DavidVaini/10308388
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

// RoundToPrecision rounds float64 to specified decimal precision.
// Source: https://gist.github.com/DavidVaini/10308388
func RoundToPrecision(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
