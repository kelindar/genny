// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package join

func JoinMyStrs(list []MyStr, sep string) (result string) {
	for i, elem := range list {
		if i > 0 {
			result += sep
		}
		result += elem.String()
	}
	return
}
