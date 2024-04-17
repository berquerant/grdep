package grdep_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

func TestReaderCategorySelector(t *testing.T) {
	for _, tc := range []struct {
		name    string
		r       io.Reader
		matcher grdep.MatcherIface
		want    []string
		err     error
	}{
		{
			name: "empty reader",
			r:    bytes.NewBufferString(""),
			matcher: MockMatcherFunc(func() ([]string, error) {
				return []string{"a"}, nil
			}),
			err: grdep.ErrUnmatched,
		},
		{
			name: "unmatched",
			r:    bytes.NewBufferString("a\nb"),
			matcher: MockMatcherFunc(func() ([]string, error) {
				return nil, grdep.ErrUnmatched
			}),
			err: grdep.ErrUnmatched,
		},
		{
			name: "second matched",
			r:    bytes.NewBufferString("a\nb"),
			matcher: MockMatcherFunc(func() func() ([]string, error) {
				var i int
				return func() ([]string, error) {
					if i > 0 {
						return nil, grdep.ErrUnmatched
					}
					i++
					return []string{"matched"}, nil
				}
			}()),
			want: []string{"matched"},
		},
		{
			name: "first matched",
			r:    bytes.NewBufferString("a\nb"),
			matcher: MockMatcherFunc(func() ([]string, error) {
				return []string{"matched"}, nil
			}),
			want: []string{"matched"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := grdep.NewReaderCategorySelector(tc.matcher)
			defer s.Close()
			got, err := s.Select(tc.r)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
