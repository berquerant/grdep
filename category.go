package grdep

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

type CategorySelectorIface interface {
	Select(path string) ([]string, error)
	Close() error
}

var (
	_ CategorySelectorIface = &FileCategorySelector{}
	_ CategorySelectorIface = &TextCategorySelector{}
)

func NewFileCategorySelector(matcher MatcherIface) CategorySelectorIface {
	return &FileCategorySelector{
		matcher: matcher,
	}
}

func NewTextCategorySelector(reader *ReaderCategorySelector) CategorySelectorIface {
	return &TextCategorySelector{
		reader: reader,
	}
}

type FileCategorySelector struct {
	matcher MatcherIface
}

func (c FileCategorySelector) Close() error {
	if c.matcher == nil {
		return nil
	}
	return c.matcher.Close()
}

func (c FileCategorySelector) Select(path string) ([]string, error) {
	r, err := c.matcher.Match(path)
	if err != nil {
		return nil, fmt.Errorf("%w: file category %s", err, path)
	}
	return r, nil
}

type TextCategorySelector struct {
	reader *ReaderCategorySelector
}

func (s *TextCategorySelector) Close() error {
	return s.reader.Close()
}

func (c TextCategorySelector) Select(path string) ([]string, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: text category %s", err, path)
	}
	defer fp.Close()

	rs, err := c.reader.Select(fp)
	if err != nil {
		return nil, fmt.Errorf("%w: text category %s", err, path)
	}
	return rs, nil
}

func NewReaderCategorySelector(matcher MatcherIface) *ReaderCategorySelector {
	return &ReaderCategorySelector{
		matcher: matcher,
	}
}

type ReaderCategorySelector struct {
	matcher MatcherIface
}

func (s ReaderCategorySelector) Close() error {
	return s.matcher.Close()
}

func (s ReaderCategorySelector) Select(r io.Reader) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for x := range ReadLines(ctx, r) {
		if err := x.Err; err != nil {
			return nil, fmt.Errorf("%w: reader category", err)
		}
		r, err := s.matcher.Match(x.Text)
		if err == nil {
			return r, nil
		}
		if errors.Is(err, ErrUnmatched) {
			continue
		}
		return nil, fmt.Errorf("%w: reader category %s", err, x)
	}

	return nil, fmt.Errorf("%w: reader category", ErrUnmatched)
}
