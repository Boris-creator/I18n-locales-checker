package selectors

import (
	"regexp"
	"strings"
)

type Selector struct {
	Selector string
}

func (s *Selector) Chunks() []string {
	return strings.Split(s.Selector, "/")
}
func From(chunks []string) Selector {
	return Selector{
		Selector: strings.Join(chunks, "/"),
	}
}

func Satisfies(target, chunk string) bool {
	if strings.Contains(chunk, "*") {
		re := regexp.MustCompile(strings.Replace(regexp.QuoteMeta(chunk), "*", ".*", -1))
		return re.MatchString(target)
	}
	return target == chunk
}
