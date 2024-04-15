package grdep

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type WalkerIface interface {
	// Walk walks the file tree and read each lines of files.
	Walk(ctx context.Context) <-chan Line
}

var (
	_ WalkerIface = &Walker{}
)

type Line struct {
	Linum   int    `json:"linum"`
	Content string `json:"content"`
	Path    string `json:"path"`
	Err     error  `json:"err,omitempty"`
}

func (r Line) String() string {
	return fmt.Sprintf("at %s:%d:%s", r.Path, r.Linum, r.Content)
}

func NewWalker(root string, ignores MatcherIface) *Walker {
	return &Walker{
		root:    root,
		ignores: ignores,
	}
}

type Walker struct {
	root    string
	ignores MatcherIface
}

func (w Walker) isSkip(path string) bool {
	_, err := w.ignores.Match(path)
	return err == nil
}

func (w Walker) Walk(ctx context.Context) <-chan Line {
	resultC := make(chan Line, 1000)

	go func() {
		defer close(resultC)

		_ = filepath.Walk(w.root, func(path string, info fs.FileInfo, walkErr error) error {
			if IsDone(ctx) {
				return ctx.Err()
			}
			if walkErr != nil {
				return nil
			}
			if w.isSkip(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.IsDir() {
				return nil
			}

			return w.scan(ctx, path, resultC)
		})
	}()

	return resultC
}

func (Walker) scan(ctx context.Context, path string, resultC chan<- Line) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	for x := range ReadLines(ctx, fp) {
		resultC <- Line{
			Linum:   x.Linum,
			Content: x.Text,
			Path:    path,
			Err:     x.Err,
		}
	}
	return nil
}
