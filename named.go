package grdep

import "fmt"

type Named interface {
	GetName() string
}

var (
	_ MatcherIface = &NamedMatcher{}
	_ Named        = &NamedMatcher{}
	_ Named        = &NamedCategorySelector{}
	_ Named        = &NamedNodeSelector{}
)

func (m NamedMatcher) GetName() string {
	return m.Name
}

func (m NamedMatcher) Match(src string) ([]string, error) {
	r, err := m.Matcher.Match(src)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, m.Name)
	}
	return r, nil
}

type NamedSelectorResult struct {
	Index  int    `json:"index"`
	Name   string `json:"name,omitempty"`
	Result string `json:"result,omitempty"`
	Err    error  `json:"err,omitempty"`
}

func NewNamedCategorySelector(name string, selector CategorySelectorIface) *NamedCategorySelector {
	return &NamedCategorySelector{
		name:     name,
		selector: selector,
	}
}

type NamedCategorySelector struct {
	name     string
	selector CategorySelectorIface
}

func (s NamedCategorySelector) GetName() string {
	return s.name
}

func (s NamedCategorySelector) Select(path string) ([]string, error) {
	category, err := s.selector.Select(path)
	if err != nil {
		return nil, fmt.Errorf("%w: category(%s)", err, s.name)
	}
	return category, nil
}

type NamedCategorySelectors []*NamedCategorySelector

func (s NamedCategorySelectors) Select(path string) []NamedSelectorResult {
	result := []NamedSelectorResult{}
	for i, x := range s {
		rs, err := x.Select(path)
		if err != nil {
			result = append(result, NamedSelectorResult{
				Index: i,
				Name:  x.name,
				Err:   err,
			})
			continue
		}
		for _, r := range rs {
			result = append(result, NamedSelectorResult{
				Index:  i,
				Name:   x.name,
				Result: r,
			})
		}
	}
	return result
}

func NewNamedNodeSelector(name string, selector NodeSelectorIface) *NamedNodeSelector {
	return &NamedNodeSelector{
		name:     name,
		selector: selector,
	}
}

type NamedNodeSelector struct {
	name     string
	selector NodeSelectorIface
}

func (s NamedNodeSelector) GetName() string {
	return s.name
}

func (s NamedNodeSelector) Select(category, content string) ([]string, error) {
	r, err := s.selector.Select(category, content)
	if err != nil {
		return nil, fmt.Errorf("%w: node(%s)", err, s.name)
	}
	return r, nil
}

type NamedNodeSelectors []*NamedNodeSelector

func (s NamedNodeSelectors) Select(category, content string) []NamedSelectorResult {
	result := []NamedSelectorResult{}
	for i, x := range s {
		rs, err := x.Select(category, content)
		if err != nil {
			result = append(result, NamedSelectorResult{
				Index: i,
				Name:  x.name,
				Err:   err,
			})
			continue
		}
		for _, r := range rs {
			result = append(result, NamedSelectorResult{
				Index:  i,
				Name:   x.name,
				Result: r,
			})
		}
	}
	return result
}

type NamedNormalizers []NamedMatcher

type NamedNormalizerResult struct {
	Index  int    `json:"index"`
	Name   string `json:"name,omitempty"`
	Result string `json:"result,omitempty"`
}

func (n NamedNormalizers) Normalize(src string) []NamedNormalizerResult {
	for idx, matcher := range n {
		if rs, err := matcher.Match(src); err == nil {
			result := make([]NamedNormalizerResult, len(rs))
			for i, r := range rs {
				result[i] = NamedNormalizerResult{
					Index:  idx,
					Name:   matcher.Name,
					Result: r,
				}
			}
			return result
		}
	}
	return []NamedNormalizerResult{
		{
			Index:  -1,
			Result: src,
		},
	}
}
