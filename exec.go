package grdep

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/berquerant/execx"
)

type ShellScript struct {
	script *execx.Script
}

func NewShellScript(content, shell string) *ShellScript {
	s := execx.NewScript(content, shell)
	s.KeepScriptFile = true
	s.Env.Merge(execx.EnvFromEnviron())
	return &ShellScript{
		script: s,
	}
}

func (s ShellScript) Run(ctx context.Context, src string) ([]string, error) {
	var result []string
	if err := s.script.Runner(func(cmd *execx.Cmd) error {
		cmd.Stdin = bytes.NewBufferString(src)
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
