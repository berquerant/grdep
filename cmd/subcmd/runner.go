package subcmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"

	"github.com/berquerant/grdep"
)

func jsonify(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

type runner struct {
	config             *grdep.Config
	r                  io.Reader
	w                  io.Writer
	logger             *slog.Logger
	isDebug            bool
	categories         func(string) []grdep.NamedSelectorResult
	nodes              func(category, content string) []grdep.NamedSelectorResult
	categoryNormalizer func(string) []grdep.NamedNormalizerResult
	nodeNormalizer     func(string) []grdep.NamedNormalizerResult
	categoryOnly       bool
}

func (r runner) debug(f func()) {
	if r.isDebug {
		f()
	}
}

func (r runner) write(v any) {
	grdep.WriteJSON(r.w, v)
}

func (r runner) run(ctx context.Context) error {
	r.debug(func() { r.logger.Debug("run") })
	for path := range grdep.ReadLines(ctx, r.r) {
		a := PassArg{
			Path: path,
		}
		if err := r.processPath(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (r runner) processPath(ctx context.Context, arg PassArg) error {
	r.debug(func() { r.logger.Debug("process path", "arg", jsonify(arg)) })
	if err := arg.Path.Err; err != nil {
		return err
	}

	for row := range grdep.NewWalker(arg.Path.Text, r.config.Ignores).Walk(ctx) {
		a := arg
		a.Line = row
		if err := r.processLine(ctx, a); err != nil {
			return err
		}
	}

	return nil
}

func (r runner) processLine(ctx context.Context, arg PassArg) error {
	r.debug(func() { r.logger.Debug("process row", "arg", jsonify(arg)) })
	if err := arg.Line.Err; err != nil {
		return err
	}

	for _, x := range r.categories(arg.Line.Path) {
		a := arg
		a.Category = x
		if err := r.processCategory(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (r runner) processCategory(ctx context.Context, arg PassArg) error {
	r.debug(func() { r.logger.Debug("process category", "arg", jsonify(arg)) })
	if errors.Is(arg.Category.Err, grdep.ErrUnmatched) {
		return nil
	}
	if err := arg.Category.Err; err != nil {
		return err
	}

	for _, x := range r.categoryNormalizer(arg.Category.Result) {
		a := arg
		a.NormalizedCategory = x
		if err := r.processNormalizedCategory(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (r runner) processNormalizedCategory(ctx context.Context, arg PassArg) error {
	r.debug(func() { r.logger.Debug("process normalized category", "arg", jsonify(arg)) })
	if r.categoryOnly {
		r.write(arg.intoResult())
		return nil
	}

	for _, x := range r.nodes(arg.NormalizedCategory.Result, arg.Line.Content) {
		a := arg
		a.Node = x
		if err := r.processNode(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (r runner) processNode(ctx context.Context, arg PassArg) error {
	r.debug(func() { r.logger.Debug("process node", "arg", jsonify(arg)) })
	if errors.Is(arg.Node.Err, grdep.ErrUnmatched) {
		return nil
	}
	if err := arg.Node.Err; err != nil {
		return err
	}

	for _, x := range r.nodeNormalizer(arg.Node.Result) {
		a := arg
		a.NormalizedNode = x
		if err := r.processNormalizedNode(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (r runner) processNormalizedNode(_ context.Context, arg PassArg) error {
	r.debug(func() { r.logger.Debug("process normalized node", "arg", jsonify(arg)) })

	r.write(arg.intoResult())
	return nil
}
