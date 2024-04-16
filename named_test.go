package grdep_test

import (
	"fmt"
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

type MockCategorySelectorFunc func() ([]string, error)

func (f MockCategorySelectorFunc) Select(_ string) ([]string, error) {
	return f()
}

func TestNamedCategorySelectors(t *testing.T) {
	for _, tc := range []struct {
		name      string
		selectors grdep.NamedCategorySelectors
		want      []grdep.NamedSelectorResult
	}{
		{
			name:      "no selectors",
			selectors: grdep.NamedCategorySelectors([]*grdep.NamedCategorySelector{}),
			want:      []grdep.NamedSelectorResult{},
		},
		{
			name: "err",
			selectors: grdep.NamedCategorySelectors([]*grdep.NamedCategorySelector{
				grdep.NewNamedCategorySelector("a", MockCategorySelectorFunc(func() ([]string, error) {
					return nil, grdep.ErrUnmatched
				})),
			}),
			want: []grdep.NamedSelectorResult{
				{
					Index: 0,
					Name:  "a",
					Err:   fmt.Errorf("%w: category(a)", grdep.ErrUnmatched),
				},
			},
		},
		{
			name: "categories",
			selectors: grdep.NamedCategorySelectors([]*grdep.NamedCategorySelector{
				grdep.NewNamedCategorySelector("a", MockCategorySelectorFunc(func() ([]string, error) {
					return []string{
						"r1",
						"r2",
					}, nil
				})),
			}),
			want: []grdep.NamedSelectorResult{
				{
					Index:  0,
					Name:   "a",
					Result: "r1",
				},
				{
					Index:  0,
					Name:   "a",
					Result: "r2",
				},
			},
		},
		{
			name: "categories and err",
			selectors: grdep.NamedCategorySelectors([]*grdep.NamedCategorySelector{
				grdep.NewNamedCategorySelector("a", MockCategorySelectorFunc(func() ([]string, error) {
					return nil, grdep.ErrUnmatched
				})),
				grdep.NewNamedCategorySelector("b", MockCategorySelectorFunc(func() ([]string, error) {
					return []string{
						"r1",
						"r2",
					}, nil
				})),
			}),
			want: []grdep.NamedSelectorResult{
				{
					Index: 0,
					Name:  "a",
					Err:   fmt.Errorf("%w: category(a)", grdep.ErrUnmatched),
				},
				{
					Index:  1,
					Name:   "b",
					Result: "r1",
				},
				{
					Index:  1,
					Name:   "b",
					Result: "r2",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.selectors.Select("")
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNamedNormalizers(t *testing.T) {
	newRegexp := func(pattern string) *grdep.Regexp {
		v := grdep.NewRegexp(pattern)
		return &v
	}
	for _, tc := range []struct {
		name        string
		normalizers grdep.NamedNormalizers
		src         string
		want        []grdep.NamedNormalizerResult
	}{
		{
			name:        "no normalizers",
			normalizers: grdep.NamedNormalizers([]grdep.NamedMatcher{}),
			src:         "in",
			want: []grdep.NamedNormalizerResult{
				{
					Index:  -1,
					Result: "in",
				},
			},
		},
		{
			name: "not matched",
			normalizers: grdep.NamedNormalizers([]grdep.NamedMatcher{
				{
					Name: "m1",
					Matcher: []*grdep.Matcher{
						{
							Regex: newRegexp(`^sh$`),
							Value: []string{"bash"},
						},
					},
				},
			}),
			src: "in",
			want: []grdep.NamedNormalizerResult{
				{
					Index:  -1,
					Result: "in",
				},
			},
		},
		{
			name: "matched",
			normalizers: grdep.NamedNormalizers([]grdep.NamedMatcher{
				{
					Name: "m1",
					Matcher: []*grdep.Matcher{
						{
							Regex: newRegexp(`^sh$`),
						},
						{
							Value: []string{"bash"},
						},
					},
				},
			}),
			src: "sh",
			want: []grdep.NamedNormalizerResult{
				{
					Index:  0,
					Name:   "m1",
					Result: "bash",
				},
			},
		},
		{
			name: "first matched",
			normalizers: grdep.NamedNormalizers([]grdep.NamedMatcher{
				{
					Name: "m1",
					Matcher: []*grdep.Matcher{
						{
							Regex: newRegexp(`^sh$`),
						},
						{
							Value: []string{"bash"},
						},
					},
				},
				{
					Name: "m2",
					Matcher: []*grdep.Matcher{
						{
							Regex: newRegexp(`^bash$`),
						},
						{
							Value: []string{"zsh"},
						},
					},
				},
			}),
			src: "sh",
			want: []grdep.NamedNormalizerResult{
				{
					Index:  0,
					Name:   "m1",
					Result: "bash",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.normalizers.Normalize(tc.src)
			assert.Equal(t, tc.want, got)
		})
	}
}
