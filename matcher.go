package grdep

import (
	"errors"
	"fmt"
)

type MatcherIface interface {
	Match(src string) ([]string, error)
}

var (
	_ MatcherIface = &Matcher{}

	ErrUnmatched = errors.New("Unmatched")
)

func (m MatchExpr) Match(src string) bool {
	if m.Not != nil {
		return !m.Not.Unwrap().MatchString(src)
	}
	return m.Regex.Unwrap().MatchString(src)
}

func (m Matcher) Match(src string) (ret []string, err error) {
	if err = m.matchHeads(src); err != nil {
		return
	}

	switch {
	case m.Template != "":
		// If match, apply template
		ret, err = m.expand(src)
	case len(m.Value) > 0:
		// If match, return the value
		ret, err = m.match(src)
	default:
		// If match, return all matched values
		ret, err = m.matchAll(src)
	}

	if err != nil {
		err = fmt.Errorf("%w: %v unmatched with %s", err, m.Regex, src)
	}
	return
}

func (m Matcher) heads() []MatchExpr {
	return m.Regex[:len(m.Regex)-1]
}

func (m Matcher) matchHeads(src string) error {
	for i, x := range m.heads() {
		if !x.Match(src) {
			return fmt.Errorf("%w: heads matcher[%d]", ErrUnmatched, i)
		}
	}
	return nil
}

func (m Matcher) tail() *Regexp {
	return m.Regex[len(m.Regex)-1].Regex
}

func (m Matcher) matchAll(src string) ([]string, error) {
	if matches := m.tail().Unwrap().FindAllString(src, -1); len(matches) > 0 {
		return matches, nil
	}
	return nil, ErrUnmatched
}

func (m Matcher) match(src string) ([]string, error) {
	if m.tail().Unwrap().MatchString(src) {
		return m.Value, nil
	}
	return nil, ErrUnmatched
}

func (m Matcher) expand(src string) ([]string, error) {
	result := []byte{}
	for _, submatches := range m.tail().Unwrap().FindAllStringSubmatchIndex(src, -1) {
		result = m.tail().Unwrap().ExpandString(result, m.Template, src, submatches)
	}
	if len(result) == 0 {
		return nil, ErrUnmatched
	}
	return []string{string(result)}, nil
}
