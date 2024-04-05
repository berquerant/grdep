package grdep

type NodeSelectorIface interface {
	Select(category, content string) ([]string, error)
}

var (
	_ NodeSelectorIface = &NodeSelector{}
)

func NewNodeSelector(category Regexp, selector MatcherIface) *NodeSelector {
	return &NodeSelector{
		category: category,
		selector: selector,
	}
}

type NodeSelector struct {
	category Regexp
	selector MatcherIface
}

func (n NodeSelector) Select(category, content string) ([]string, error) {
	if !n.category.Unwrap().MatchString(category) {
		return nil, ErrUnmatched
	}
	r, err := n.selector.Match(content)
	if err != nil {
		return nil, err
	}
	return r, nil
}
