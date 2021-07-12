// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package syntax

import "time"

type timeSpanList []time.Duration

type TimeSpanUppercase time.Duration

var _ time.Duration
var timeSpanVariable string

func _() {
	var _ []time.Duration // A comment
	var _ timeSpanList
	var _ []timeSpanList
}

func PrintTimeSpan(_timeSpan time.Duration) {
	var _ TimeSpanUppercase
	var u interface{}

	v := u.(TimeSpanUppercase)
	println(timeSpanVariable, _timeSpan, time.Duration(123), v)
}

type fractionalList []float64

type FractionalUppercase float64

var _ float64
var fractionalVariable string

func _() {
	var _ []float64 // A comment
	var _ fractionalList
	var _ []fractionalList
}

func PrintFractional(_fractional float64) {
	var _ FractionalUppercase
	var u interface{}

	v := u.(FractionalUppercase)
	println(fractionalVariable, _fractional, float64(123), v)
}
