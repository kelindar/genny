package example

import "github.com/kelindar/genny/generic"

type number = generic.Number

// Number returns the number
func Number() number {
	return number(0)
}
