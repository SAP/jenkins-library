package piperutils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Title(in string) string {
	return cases.Title(language.English, cases.NoLower).String(in)
}
