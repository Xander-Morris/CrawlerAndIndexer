package utils

import (
	"github.com/microcosm-cc/bluemonday"
)

func StripHTML(input string) string {
	p := bluemonday.StrictPolicy()
	return p.Sanitize(input)
}
