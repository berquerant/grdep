package grdep

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type ReadLinesResult struct {
	Linum int    `json:"linum"`
	Text  string `json:"text"`
	Err   error  `json:"err,omitempty"`
}

func (r ReadLinesResult) String() string {
	return fmt.Sprintf("%d:%s", r.Linum, r.Text)
}

func ReadLines(ctx context.Context, r io.Reader) <-chan ReadLinesResult {
	resultC := make(chan ReadLinesResult, 1000)

	go func() {
		defer close(resultC)

		var (
			scanner = bufio.NewScanner(r)
			linum   int
		)
		for scanner.Scan() {
			if IsDone(ctx) {
				resultC <- ReadLinesResult{
					Err: ctx.Err(),
				}
				return
			}

			linum++
			resultC <- ReadLinesResult{
				Linum: linum,
				Text:  scanner.Text(),
			}
		}
		if err := scanner.Err(); err != nil {
			resultC <- ReadLinesResult{
				Err: err,
			}
		}
	}()

	return resultC
}

func WriteJSON(w io.Writer, v any) {
	b, _ := json.Marshal(v)
	fmt.Fprintln(w, string(b))
}
