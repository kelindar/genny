// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package renamed

import (
	"fmt"

	testpkg "github.com/kelindar/genny/parse/test/renamed/subpkg"
)

func someFuncint() {
	var t int
	fmt.Println(t)
	fmt.Println(testpkg.Bar)
}
