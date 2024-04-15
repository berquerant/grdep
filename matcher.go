package grdep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/berquerant/execx"
)

type MatcherSet []*Matcher

func (m MatcherSet) Match(src string) ([]string, error) {
	if len(m) == 0 {
		return nil, ErrUnmatched
	}

	result := []string{src}
	for i, x := range m {
		acc := []string{}
		for _, y := range result {
			r, err := x.Match(y)
			OnDebug(func() {
				b, _ := json.Marshal(x)
				L().Debug("matcher", "index", i, "body", string(b), "src", y, "ret", r, "err", err)
			})
			if err != nil {
				continue
			}
			acc = append(acc, r...)
		}
		if len(acc) == 0 {
			return nil, fmt.Errorf("%w: matcher set[%d]", ErrUnmatched, i)
		}
		result = acc
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: matcher set", ErrUnmatched)
	}
	return result, nil
}

type MatcherIface interface {
	Match(src string) ([]string, error)
}

var (
	_    MatcherIface = &Matcher{}
	__ms MatcherSet
	_    MatcherIface = &__ms

	ErrUnmatched = errors.New("Unmatched")
)

func (m *Matcher) Match(src string) ([]string, error) {
	r, err := m.internalMatch(src)
	if err != nil {
		return nil, err
	}
	r = slices.DeleteFunc(r, func(x string) bool {
		return strings.TrimSpace(x) == ""
	})
	if len(r) == 0 {
		return nil, ErrUnmatched
	}
	return r, nil
}

func (m *Matcher) internalMatch(src string) ([]string, error) {
	switch {
	case m.Shell != "":
		return m.runShell(src)
	case m.Not != nil:
		return m.notMatch(src)
	case m.Template != "":
		return m.expand(src)
	case m.Regex != nil:
		return m.match(src)
	case len(m.Value) > 0:
		return m.value(src)
	default:
		return nil, ErrUnmatched
	}
}

func (m *Matcher) value(_ string) ([]string, error) {
	return m.Value, nil
}

func (m *Matcher) notMatch(src string) ([]string, error) {
	if !m.Not.Unwrap().MatchString(src) {
		if len(m.Value) > 0 {
			return m.Value, nil
		}
		return []string{src}, nil
	}
	return nil, ErrUnmatched
}

func (m *Matcher) match(src string) ([]string, error) {
	if m.Regex.Unwrap().MatchString(src) {
		if len(m.Value) > 0 {
			return m.Value, nil
		}
		return []string{src}, nil
	}
	return nil, ErrUnmatched
}

func (m *Matcher) expand(src string) ([]string, error) {
	result := []byte{}
	for _, submatches := range m.Regex.Unwrap().FindAllStringSubmatchIndex(src, -1) {
		result = m.Regex.Unwrap().ExpandString(result, m.Template, src, submatches)
	}
	if len(result) == 0 {
		return nil, ErrUnmatched
	}
	return []string{string(result)}, nil
}

func (m *Matcher) prepareShell() {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.shellScript != nil {
		return
	}
	m.shellScript = execx.NewScript(m.Shell, "bash")
	m.shellScript.KeepScriptFile = true
	m.shellScript.Env.Merge(execx.EnvFromEnviron())
}

func (m *Matcher) runShell(src string) ([]string, error) {
	r, err := m.internalRunShell(src)
	if err != nil {
		return nil, errors.Join(ErrUnmatched, err)
	}
	return r, nil
}

func (m *Matcher) internalRunShell(src string) ([]string, error) {
	m.prepareShell()

	var result []string
	if err := m.shellScript.Runner(func(cmd *execx.Cmd) error {
		cmd.Stdin = bytes.NewBufferString(src)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := cmd.Run(ctx)
		if err != nil {
			return err
		}
		b, err := io.ReadAll(r.Stdout)
		if err != nil {
			return err
		}
		result = strings.Split(string(b), "\n")
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}
