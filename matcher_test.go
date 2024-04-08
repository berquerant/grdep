package grdep_test

import (
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

func TestMatcher(t *testing.T) {
	newRegexp := func(pattern string) *grdep.Regexp {
		v := grdep.NewRegexp(pattern)
		return &v
	}
	for _, tc := range []struct {
		name    string
		matcher *grdep.Matcher
		src     string
		want    []string
		err     error
	}{
		{
			name: "value matched",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`target`),
					},
				},
				Value: []string{"ret"},
			},
			src:  "target",
			want: []string{"ret"},
		},
		{
			name: "value unmatched",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`target`),
					},
				},
				Value: []string{"ret"},
			},
			src: "unmatched",
			err: grdep.ErrUnmatched,
		},
		{
			name: "template matched",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`/(?P<sh>[^/]+)$`),
					},
				},
				Template: "$sh",
			},
			src:  "#!/bin/bash",
			want: []string{"bash"},
		},
		{
			name: "template unmatched",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`/(?P<sh>[^/]+)$`),
					},
				},
				Template: "$sh",
			},
			src: "func()",
			err: grdep.ErrUnmatched,
		},
		{
			name: "match all",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`/bin/[^,]+`),
					},
				},
			},
			src: "/bin/a,/bin/b",
			want: []string{
				"/bin/a",
				"/bin/b",
			},
		},
		{
			name: "match all all",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`bin`),
					},
					{
						Regex: newRegexp(`/bin/[^,]+`),
					},
				},
			},
			src: "/bin/a,/bin/b",
			want: []string{
				"/bin/a",
				"/bin/b",
			},
		},
		{
			name: "unmatched at first",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`^/sbin`),
					},
					{
						Regex: newRegexp(`/bin/[^,]+`),
					},
				},
			},
			src: "/bin/a,/bin/b",
			err: grdep.ErrUnmatched,
		},
		{
			name: "template matched all",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`bash`),
					},
					{
						Regex: newRegexp(`/(?P<sh>[^/]+)$`),
					},
				},
				Template: "$sh",
			},
			src:  "#!/bin/bash",
			want: []string{"bash"},
		},
		{
			name: "value matched all",
			matcher: &grdep.Matcher{
				Regex: []grdep.MatchExpr{
					{
						Regex: newRegexp(`target`),
					},
					{
						Regex: newRegexp(`tar`),
					},
				},
				Value: []string{"ret"},
			},
			src:  "target",
			want: []string{"ret"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.matcher.Match(tc.src)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

type MockMatcherFunc func() ([]string, error)

func (f MockMatcherFunc) Match(_ string) ([]string, error) {
	return f()
}
