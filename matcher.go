package grdep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type NamedMatcherSet struct {
	MatcherSet
	name string
}

func NewNamedMatcherSet(name string, matcherSet MatcherSet) *NamedMatcherSet {
	return &NamedMatcherSet{
		name:       name,
		MatcherSet: matcherSet,
	}
}

func (m NamedMatcherSet) Match(src string) ([]string, error) {
	return AddMetric(fmt.Sprintf("named-matcher-set-%s", m.name), func() ([]string, error) {
		return m.MatcherSet.Match(src)
	})
}

type MatcherSet []*Matcher

func (m MatcherSet) Close() error {
	for _, x := range m {
		_ = x.Close()
	}
	return nil
}

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
	Close() error
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
	case m.LuaEntryPoint != "":
		return AddMetric("matcher-lua", func() ([]string, error) {
			return m.runLua(src)
		})
	case m.Shell != "":
		return AddMetric("matcher-shell", func() ([]string, error) {
			return m.runShell(src)
		})
	case m.Glob != "":
		return AddMetric("matcher-glob", func() ([]string, error) {
			return m.glob(src)
		})
	case m.Not != nil:
		return AddMetric("matcher-not", func() ([]string, error) {
			return m.notMatch(src)
		})
	case m.Template != "":
		return AddMetric("matcher-template", func() ([]string, error) {
			return m.expand(src)
		})
	case m.Regex != nil:
		return AddMetric("matcher-regex", func() ([]string, error) {
			return m.match(src)
		})
	case len(m.Value) > 0:
		return AddMetric("matcher-value", func() ([]string, error) {
			return m.value(src)
		})
	default:
		return nil, ErrUnmatched
	}
}

func (m *Matcher) Close() error {
	if m == nil {
		return nil
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	if m.luaScript != nil {
		m.luaScript.Close()
	}
	return nil
}

func (m *Matcher) value(_ string) ([]string, error) {
	return m.Value, nil
}

func (m *Matcher) notMatch(src string) ([]string, error) {
	if !m.Not.Unwrap().MatchString(src) {
		return []string{src}, nil
	}
	return nil, ErrUnmatched
}

func (m *Matcher) match(src string) ([]string, error) {
	if m.Regex.Unwrap().MatchString(src) {
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

const (
	shellScriptTimeout = 3 * time.Second
)

func (m *Matcher) prepareShell() {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.shellScript != nil {
		return
	}
	m.shellScript = NewShellScript(m.Shell, "bash")
}

func (m *Matcher) runShell(src string) ([]string, error) {
	r, err := m.internalRunShell(src)
	if err != nil {
		return nil, errors.Join(ErrUnmatched, err)
	}
	return r, nil
}

func (m *Matcher) internalRunShell(src string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), shellScriptTimeout)
	defer cancel()

	m.prepareShell()
	return m.shellScript.Run(ctx, src)
}

func (m *Matcher) glob(src string) ([]string, error) {
	matched, err := filepath.Match(m.Glob, src)
	if err != nil {
		return nil, errors.Join(ErrUnmatched, err)
	}
	if !matched {
		return nil, ErrUnmatched
	}
	return []string{src}, nil
}

func (m *Matcher) prepareLua() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.luaScript != nil {
		return nil
	}
	var (
		s   *LuaScript
		err error
	)
	if m.LuaFile != "" {
		s, err = NewLuaScriptFromFile(m.LuaFile, m.LuaEntryPoint)
	} else {
		s, err = NewLuaScript(m.Lua, m.LuaEntryPoint)
	}
	if err != nil {
		return err
	}
	m.luaScript = s
	return nil
}

func (m *Matcher) runLua(src string) ([]string, error) {
	r, err := m.internalRunLua(src)
	if err != nil {
		return nil, errors.Join(ErrUnmatched, err)
	}
	return r, nil
}

func (m *Matcher) internalRunLua(src string) ([]string, error) {
	if err := m.prepareLua(); err != nil {
		return nil, err
	}
	return m.luaScript.Run(src)
}
