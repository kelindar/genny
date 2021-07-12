// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package bugreports

type DigraphInt struct {
	ints map[int][]int
}

func NewDigraphInt() *DigraphInt {
	return &DigraphInt{
		ints: make(map[int][]int),
	}
}

func (dig *DigraphInt) Add(n int) {
	if _, exists := dig.ints[n]; exists {
		return
	}

	dig.ints[n] = nil
}

func (dig *DigraphInt) Connect(a, b int) {
	dig.Add(a)
	dig.Add(b)

	dig.ints[a] = append(dig.ints[a], b)
}
