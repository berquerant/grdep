package grdep

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type LuaScript struct {
	entryPoint string // call function: string -> string

	state *lua.LState
	mux   sync.Mutex
}

func NewLuaScriptFromFile(script, entryPoint string) (*LuaScript, error) {
	state := lua.NewState()
	if err := state.DoFile(script); err != nil {
		return nil, errors.Join(ErrLuaInvalidScript, err)
	}

	return &LuaScript{
		entryPoint: entryPoint,
		state:      state,
	}, nil
}

func NewLuaScript(script, entryPoint string) (*LuaScript, error) {
	state := lua.NewState()
	if err := state.DoString(script); err != nil {
		return nil, errors.Join(ErrLuaInvalidScript, err)
	}

	return &LuaScript{
		entryPoint: entryPoint,
		state:      state,
	}, nil
}

func (s *LuaScript) Close() {
	s.state.Close()
}

var (
	ErrLuaInvalidScript     = errors.New("LuaInvalidScript")
	ErrLuaInvalidCall       = errors.New("LuaInvalidCall")
	ErrLuaInvalidReturnType = errors.New("LuaInvalidReturnType")
)

func (s *LuaScript) Run(src string) ([]string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal(s.entryPoint),
		NRet:    1,
		Protect: true,
	}, lua.LString(src)); err != nil {
		return nil, errors.Join(ErrLuaInvalidCall, err)
	}

	lRet := s.state.Get(-1)
	s.state.Pop(1)

	lStr, ok := lRet.(lua.LString)
	if !ok {
		return nil, fmt.Errorf("%w: return type %s but should be String", ErrLuaInvalidReturnType, lRet.Type())
	}

	result := strings.Split(lStr.String(), "\n")
	return result, nil
}
