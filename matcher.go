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

func (m Matcher) Match(src string) (ret []string, err error) {
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

func (m Matcher) matchAll(src string) ([]string, error) {
	if matches := m.Regex.Unwrap().FindAllString(src, -1); len(matches) > 0 {
		return matches, nil
	}
	return nil, ErrUnmatched
}

func (m Matcher) match(src string) ([]string, error) {
	if m.Regex.Unwrap().MatchString(src) {
		return m.Value, nil
	}
	return nil, ErrUnmatched
}

func (m Matcher) expand(src string) ([]string, error) {
	result := []byte{}
	for _, submatches := range m.Regex.Unwrap().FindAllStringSubmatchIndex(src, -1) {
		result = m.Regex.Unwrap().ExpandString(result, m.Template, src, submatches)
	}
	if len(result) == 0 {
		return nil, ErrUnmatched
	}
	return []string{string(result)}, nil
}
