package grdep_test

import (
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		var (
			zero grdep.Config
			r1   = grdep.NewRegexp(`r1`)
			r2   = grdep.NewRegexp(`r2`)
			c1   = grdep.CSelector{
				Name: "c1",
			}
			c2 = grdep.CSelector{
				Name: "c2",
			}
			n1 = grdep.NSelector{
				Name: "n1",
			}
			n2 = grdep.NSelector{
				Name: "n2",
			}
			nc1 = grdep.NamedMatcher{
				Name: "nc1",
			}
			nc2 = grdep.NamedMatcher{
				Name: "nc2",
			}
			nn1 = grdep.NamedMatcher{
				Name: "nn1",
			}
			nn2 = grdep.NamedMatcher{
				Name: "nn2",
			}
			nr1 = grdep.Normalizers{
				Categories: []grdep.NamedMatcher{nc1},
				Nodes:      []grdep.NamedMatcher{nn1},
			}
			nr2 = grdep.Normalizers{
				Categories: []grdep.NamedMatcher{nc2},
				Nodes:      []grdep.NamedMatcher{nn2},
			}
		)

		for _, tc := range []struct {
			name  string
			left  grdep.Config
			right grdep.Config
			want  grdep.Config
		}{
			{
				name: "zero",
				want: zero,
			},
			{
				name: "right zero",
				left: grdep.Config{
					Ignores:     []grdep.Regexp{r1},
					Categories:  []grdep.CSelector{c1},
					Nodes:       []grdep.NSelector{n1},
					Normalizers: nr1,
				},
				right: zero,
				want: grdep.Config{
					Ignores:     []grdep.Regexp{r1},
					Categories:  []grdep.CSelector{c1},
					Nodes:       []grdep.NSelector{n1},
					Normalizers: nr1,
				},
			},
			{
				name: "left zero",
				left: zero,
				right: grdep.Config{
					Ignores:     []grdep.Regexp{r1},
					Categories:  []grdep.CSelector{c1},
					Nodes:       []grdep.NSelector{n1},
					Normalizers: nr1,
				},
				want: grdep.Config{
					Ignores:     []grdep.Regexp{r1},
					Categories:  []grdep.CSelector{c1},
					Nodes:       []grdep.NSelector{n1},
					Normalizers: nr1,
				},
			},
			{
				name: "add",
				left: grdep.Config{
					Ignores:     []grdep.Regexp{r1},
					Categories:  []grdep.CSelector{c1},
					Nodes:       []grdep.NSelector{n1},
					Normalizers: nr1,
				},
				right: grdep.Config{
					Ignores:     []grdep.Regexp{r2},
					Categories:  []grdep.CSelector{c2},
					Nodes:       []grdep.NSelector{n2},
					Normalizers: nr2,
				},
				want: grdep.Config{
					Ignores:    []grdep.Regexp{r1, r2},
					Categories: []grdep.CSelector{c1, c2},
					Nodes:      []grdep.NSelector{n1, n2},
					Normalizers: grdep.Normalizers{
						Categories: []grdep.NamedMatcher{nc1, nc2},
						Nodes:      []grdep.NamedMatcher{nn1, nn2},
					},
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got := tc.left.Add(tc.right)
				assert.Equal(t, tc.want, got)
			})
		}
	})
}
