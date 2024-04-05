package grdep_test

import (
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

func TestMatcher(t *testing.T) {
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
				Regex: grdep.NewRegexp(`target`),
				Value: []string{"ret"},
			},
			src:  "target",
			want: []string{"ret"},
		},
		{
			name: "value unmatched",
			matcher: &grdep.Matcher{
				Regex: grdep.NewRegexp(`target`),
				Value: []string{"ret"},
			},
			src: "unmatched",
			err: grdep.ErrUnmatched,
		},
		{
			name: "template matched",
			matcher: &grdep.Matcher{
				Regex:    grdep.NewRegexp(`/(?P<sh>[^/]+)$`),
				Template: "$sh",
			},
			src:  "#!/bin/bash",
			want: []string{"bash"},
		},
		{
			name: "template unmatched",
			matcher: &grdep.Matcher{
				Regex:    grdep.NewRegexp(`/(?P<sh>[^/]+)$`),
				Template: "$sh",
			},
			src: "func()",
			err: grdep.ErrUnmatched,
		},
		{
			name: "match all",
			matcher: &grdep.Matcher{
				Regex: grdep.NewRegexp(`/bin/[^,]+`),
			},
			src: "/bin/a,/bin/b",
			want: []string{
				"/bin/a",
				"/bin/b",
			},
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
