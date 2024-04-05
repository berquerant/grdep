package grdep

import "sync"

// CachedFunc creates a function with in-memory cache using arguments as keys.
func CachedFunc[I comparable, O any](f func(I) O) func(I) O {
	var (
		m   = make(map[I]O)
		mux sync.Mutex
	)
	return func(arg I) O {
		mux.Lock()
		defer mux.Unlock()
		v, exist := m[arg]
		if exist {
			return v
		}
		r := f(arg)
		m[arg] = r
		return r
	}
}
