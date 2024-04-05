package grdep_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/berquerant/grdep"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestCachedFunc(t *testing.T) {
	const (
		concurrency = 4
		seedSize    = 100
		delay       = 10 * time.Millisecond
	)
	var (
		called   uint32
		identity = func(x int) int {
			time.Sleep(delay)
			atomic.AddUint32(&called, 1)
			return x
		}
		target    = grdep.CachedFunc(identity)
		errResult = errors.New("Result")
		eg        errgroup.Group
	)
	for j := 0; j < concurrency; j++ {
		reversed := j%2 == 1
		if reversed {
			eg.Go(func() error {
				for i := seedSize - 1; i >= 0; i-- {
					if i != target(i) {
						return errResult
					}
				}
				return nil
			})
			continue
		}
		eg.Go(func() error {
			for i := 0; i < seedSize; i++ {
				if i != target(i) {
					return errResult
				}
			}
			return nil
		})
	}
	assert.Nil(t, eg.Wait())
	assert.Equal(t, seedSize, int(called))
}
