package notes

import (
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

type rule struct {
	*regexp.Regexp
	action func(*http.Client, string, string) (string, error)
}

var rules []rule

func init() {
	rules = []rule{
		{Regexp: regexp.MustCompile(`(?i)^\s*inspire\s+me\s*$`), action: inspireMe},
		{Regexp: regexp.MustCompile(`(?i)^\s*remind\s+me\s*`), action: remind},
		{Regexp: regexp.MustCompile(``), action: todo},
	}
}

func Dispatch(cl *http.Client, tok, input string) (string, error) {
	for _, r := range rules {
		if !r.MatchString(input) {
			continue
		}
		return r.action(cl, tok, input)
	}

	return "", errors.New("no rules matched")
}
