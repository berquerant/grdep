package grdep_test

import (
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

func TestNodeSelector(t *testing.T) {
	t.Run("Select", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			selector *grdep.NodeSelector
			category string
			content  string
			want     []string
			err      error
		}{
			{
				name:     "category unmatched",
				selector: grdep.NewNodeSelector(grdep.NewRegexp(`bash`), nil),
				category: "go",
				err:      grdep.ErrUnmatched,
			},
			{
				name: "content unmatched",
				selector: grdep.NewNodeSelector(
					grdep.NewRegexp(`bash`),
					MockMatcherFunc(func() ([]string, error) {
						return nil, grdep.ErrUnmatched
					}),
				),
				category: "bash",
				err:      grdep.ErrUnmatched,
			},
			{
				name: "matched",
				selector: grdep.NewNodeSelector(
					grdep.NewRegexp(`bash`),
					MockMatcherFunc(func() ([]string, error) {
						return []string{"some.sh"}, nil
					}),
				),
				category: "bash",
				want:     []string{"some.sh"},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.selector.Select(tc.category, tc.content)
				if tc.err != nil {
					assert.ErrorIs(t, err, tc.err)
					return
				}
				assert.Nil(t, err)
				assert.Equal(t, tc.want, got)
			})
		}
	})
}
