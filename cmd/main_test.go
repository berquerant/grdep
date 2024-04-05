package main_test

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	based, err := os.MkdirTemp("", "e2e")
	fail(t, err)
	defer os.RemoveAll(based)
	bin := filepath.Join(based, "grdep")
	fail(t, compileBinary(bin))
	skeleton := filepath.Join(based, "skeleton.yml")

	t.Run("skeleton", func(t *testing.T) {
		f, err := os.Create(skeleton)
		fail(t, err)

		out, err := exec.Command(bin, "skeleton").Output()
		fail(t, err)
		_, err = f.Write(out)
		fail(t, err)
		fail(t, f.Close())

		t.Run("configcheck", func(t *testing.T) {
			assert.Nil(t, run(bin, "configcheck", skeleton))
		})
	})

	t.Run("run", func(t *testing.T) {
		const (
			input  = "test/target"
			golden = "test/golden.json"
		)
		config := skeleton
		tidy := func(s string) []any {
			ss := strings.Split(strings.TrimSpace(s), "\n")
			sort.Strings(ss)
			r := make([]any, len(ss))
			for i, s := range ss {
				var v any
				fail(t, json.Unmarshal([]byte(s), &v))
				r[i] = v
			}
			return r
		}

		var want []any
		{
			fp, err := os.Open(golden)
			fail(t, err)
			defer fp.Close()
			b, err := io.ReadAll(fp)
			fail(t, err)
			want = tidy(string(b))
		}

		var out strings.Builder
		cmd := exec.Command(bin, "run", config)
		cmd.Stdin = strings.NewReader(input)
		cmd.Stdout = &out
		fail(t, cmd.Run())
		got := tidy(out.String())
		assert.Equal(t, want, got)
	})
}

func compileBinary(path string) error {
	return run("go", "build", "-o", path, "-v")
}

func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fail(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
