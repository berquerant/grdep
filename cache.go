package grdep

import (
	"github.com/berquerant/cache"
)

// CachedFunc creates a function with in-memory cache using arguments as keys.
func CachedFunc[I comparable, O any](f func(I) O) func(I) O {
	c, _ := cache.NewLRU(1000, func(i I) (O, error) {
		return f(i), nil
	})

	return func(i I) O {
		v, _ := c.Get(i)
		return v
	}
}
