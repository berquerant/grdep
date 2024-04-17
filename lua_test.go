package grdep_test

import (
	"os"
	"testing"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
)

func TestLuaScript(t *testing.T) {

	t.Run("Run", func(t *testing.T) {
		t.Run("File", func(t *testing.T) {
			const (
				entryPoint = "f"
				script     = `function f(src)
  return src .. "!"
end`
			)
			fname := t.TempDir() + "/test.lua"
			if !assert.Nil(t, os.WriteFile(fname, []byte(script), 0755)) {
				return
			}

			s, err := grdep.NewLuaScriptFromFile(fname, entryPoint)
			if !assert.Nil(t, err) {
				return
			}
			got, err := s.Run("f")
			assert.Nil(t, err)
			assert.Equal(t, []string{"f!"}, got)
		})

		t.Run("Serial", func(t *testing.T) {
			const (
				entryPoint = "f"
				script     = `function f(src)
  return src .. "!"
end`
			)
			s, err := grdep.NewLuaScript(script, entryPoint)
			assert.Nil(t, err)
			defer s.Close()

			t.Run("1st", func(t *testing.T) {
				r, err := s.Run("1")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1!"}, r)
			})
			t.Run("2nd", func(t *testing.T) {
				r, err := s.Run("2")
				assert.Nil(t, err)
				assert.Equal(t, []string{"2!"}, r)
			})
		})

		t.Run("New", func(t *testing.T) {
			for _, tc := range []struct {
				name       string
				src        string
				script     string
				entryPoint string
				want       []string
				err        error
			}{
				{
					name: "invalid argument",
					src:  "a",
					script: `function g()
  return "b"
end`,
					entryPoint: "g",
					want:       []string{"b"},
				},
				{
					name: "invalid return",
					src:  "a",
					script: `function g(src)
  return 10
end`,
					entryPoint: "g",
					err:        grdep.ErrLuaInvalidReturnType,
				},
				{
					name: "hello",
					src:  "",
					script: `function f(src)
  return "hello"
end`,
					entryPoint: "f",
					want:       []string{"hello"},
				},
				{
					name: "hello name",
					src:  "name",
					script: `function f(src)
  return "hello " .. src
end`,
					entryPoint: "f",
					want:       []string{"hello name"},
				},
				{
					name: "2 lines",
					src:  "name",
					script: `function g(src)
  return "hello\n" .. src
end`,
					entryPoint: "g",
					want:       []string{"hello", "name"},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					s, err := grdep.NewLuaScript(tc.script, tc.entryPoint)
					if !assert.Nil(t, err) {
						return
					}
					got, err := s.Run(tc.src)
					if tc.err != nil {
						assert.ErrorIs(t, err, tc.err)
						return
					}
					assert.Nil(t, err)
					assert.Equal(t, got, tc.want)
				})
			}
		})
	})
}
