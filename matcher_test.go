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
		matcher grdep.MatcherIface
		src     string
		want    []string
		err     error
	}{
		{
			name: "exec sh fail",
			matcher: &grdep.Matcher{
				Shell: "grep unmatch",
			},
			src: "abc",
			err: grdep.ErrUnmatched,
		},
		{
			name: "exec sh",
			matcher: &grdep.Matcher{
				Shell: "tr ' ' '\n'",
			},
			src:  "a b c",
			want: []string{"a", "b", "c"},
		},
		{
			name: "value",
			matcher: &grdep.Matcher{
				Value: []string{"ret"},
			},
			src:  "target",
			want: []string{"ret"},
		},
		{
			name: "value matched",
			matcher: &grdep.Matcher{
				Regex: newRegexp(`target`),
				Value: []string{"ret"},
			},
			src:  "target",
			want: []string{"ret"},
		},
		{
			name: "value unmatched",
			matcher: &grdep.Matcher{
				Regex: newRegexp(`target`),
				Value: []string{"ret"},
			},
			src: "unmatched",
			err: grdep.ErrUnmatched,
		},
		{
			name: "not value matched",
			matcher: &grdep.Matcher{
				Not:   newRegexp(`not`),
				Value: []string{"ret"},
			},
			src:  "target",
			want: []string{"ret"},
		},
		{
			name: "npt value unmatched",
			matcher: &grdep.Matcher{
				Not:   newRegexp(`unmatched`),
				Value: []string{"ret"},
			},
			src: "unmatched",
			err: grdep.ErrUnmatched,
		},
		{
			name: "template matched",
			matcher: &grdep.Matcher{
				Regex:    newRegexp(`/(?P<sh>[^/]+)$`),
				Template: "$sh",
			},
			src:  "#!/bin/bash",
			want: []string{"bash"},
		},
		{
			name: "template unmatched",
			matcher: &grdep.Matcher{
				Regex:    newRegexp(`/(?P<sh>[^/]+)$`),
				Template: "$sh",
			},
			src: "func()",
			err: grdep.ErrUnmatched,
		},
		{
			name: "match all",
			matcher: &grdep.Matcher{
				Regex: newRegexp(`/bin/[^,]+`),
			},
			src:  "/bin/a,/bin/b",
			want: []string{"/bin/a,/bin/b"},
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
