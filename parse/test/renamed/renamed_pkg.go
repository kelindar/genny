package renamed

import (
	"fmt"

	"github.com/kelindar/genny/generic"

	testpkg "github.com/kelindar/genny/parse/test/renamed/subpkg"
)

type _t_ generic.Type

func someFunc_t_() {
	var t _t_
	fmt.Println(t)
	fmt.Println(testpkg.Bar)
}
