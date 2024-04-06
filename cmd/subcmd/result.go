package subcmd

import "github.com/berquerant/grdep"

type Result struct {
	Path     grdep.ReadLinesResult `json:"path,omitempty"`
	Line     grdep.Line            `json:"line,omitempty"`
	Category Selected              `json:"category,omitempty"`
	Node     Selected              `json:"node,omitempty"`
}

type Selected struct {
	Origin     grdep.NamedSelectorResult   `json:"origin,omitempty"`
	Normalized grdep.NamedNormalizerResult `json:"normalized,omitempty"`
}

type PassArg struct {
	Path               grdep.ReadLinesResult
	Line               grdep.Line
	Category           grdep.NamedSelectorResult
	NormalizedCategory grdep.NamedNormalizerResult
	Node               grdep.NamedSelectorResult
	NormalizedNode     grdep.NamedNormalizerResult
}

func (p PassArg) intoResult() Result {
	return Result{
		Path: p.Path,
		Line: p.Line,
		Category: Selected{
			Origin:     p.Category,
			Normalized: p.NormalizedCategory,
		},
		Node: Selected{
			Origin:     p.Node,
			Normalized: p.NormalizedNode,
		},
	}
}
