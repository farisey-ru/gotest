package regexp_ext

import (
	"regexp"
)

type RegexpArray []regexp.Regexp

func CompileExpr(expr []string) (*RegexpArray, error) {
	var ra = make(RegexpArray, len(expr))
	for i, s := range expr {
		x, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		ra[i] = *x
	}
	return &ra, nil
}

func (ra RegexpArray) MatchString(s string) bool {
	for _, r := range ra {
		if r.MatchString(s) {
			return true
		}
	}
	return false
}
