// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/mauricelam/genny

package syntax

type specificList []specific

type SpecificUppercase specific

var _ specific
var specificVariable string

func _() {
	var _ []specific
	var _ specificList
	var _ []specificList
}

func PrintSpecific(_specific specific) {
	var _ SpecificUppercase

	println(specificVariable, _specific)
}
