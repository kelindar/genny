// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package bugreports

// Receiver is the struct used for tests.
type Receiver struct{}

// StringsToInts converts a []string to a []int
func (Receiver) StringsToInts(string []string) []int {
	// returning an empty int slice is sufficient for this test.
	return []int{}
}

// StringsToFloat64s converts a []string to a []float64
func (Receiver) StringsToFloat64s(string []string) []float64 {
	// returning an empty float64 slice is sufficient for this test.
	return []float64{}
}

// StringsToBools converts a []string to a []bool
func (Receiver) StringsToBools(string []string) []bool {
	// returning an empty bool slice is sufficient for this test.
	return []bool{}
}
