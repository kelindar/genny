// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package main

func JoinMyStrs(list []MyStr, sep string) (result string) {
	for i, elem := range list {
		if i > 0 {
			result += sep
		}
		result += elem.String()
	}
	return
}
