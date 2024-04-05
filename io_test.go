package grdep_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	type line struct {
		linum int
		text  string
	}

	for _, tc := range []struct {
		name string
		buf  string
		want []grdep.ReadLinesResult
	}{
		{
			name: "empty",
			buf:  ``,
			want: []grdep.ReadLinesResult{},
		},
		{
			name: "a line",
			buf:  `line1`,
			want: []grdep.ReadLinesResult{
				{
					Linum: 1,
					Text:  "line1",
				},
			},
		},
		{
			name: "2 lines",
			buf: `line1
line2`,
			want: []grdep.ReadLinesResult{
				{
					Linum: 1,
					Text:  "line1",
				},
				{
					Linum: 2,
					Text:  "line2",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := []grdep.ReadLinesResult{}
			for x := range grdep.ReadLines(context.TODO(), bytes.NewBufferString(tc.buf)) {
				assert.Nil(t, x.Err)
				got = append(got, x)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
