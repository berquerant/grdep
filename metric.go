package grdep

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type GaugeMap struct {
	d   map[string]*GaugeMapElement
	mux sync.RWMutex
}

type GaugeMapElement struct {
	Count    uint64
	Duration time.Duration
}

func (g *GaugeMap) Walk(f func(key string, value *GaugeMapElement)) {
	g.mux.RLock()
	defer g.mux.RUnlock()
	for k, v := range g.d {
		f(k, v)
	}
}

func (*GaugeMap) sanitizeKey(key string) string {
	return strings.ReplaceAll(key, " ", "_")
}

func (g *GaugeMap) add(key string, count uint64, duration time.Duration) {
	key = g.sanitizeKey(key)
	if v, ok := g.d[key]; ok {
		v.Count += count
		v.Duration += duration
		return
	}

	g.d[key] = &GaugeMapElement{
		Count:    count,
		Duration: duration,
	}
}

func (g *GaugeMap) Incr(key string, f func() (any, error)) (any, error) {
	start := time.Now()
	ret, err := f()
	elapsed := time.Since(start)

	g.mux.Lock()
	defer g.mux.Unlock()

	g.add(fmt.Sprintf("%s-call", key), 1, elapsed)
	if err != nil {
		g.add(fmt.Sprintf("%s-error", key), 1, elapsed)
	} else {
		g.add(fmt.Sprintf("%s-success", key), 1, elapsed)
	}

	return ret, err
}

func NewGaugeMap() *GaugeMap {
	return &GaugeMap{
		d: map[string]*GaugeMapElement{},
	}
}

var (
	metricGaugeMap = NewGaugeMap()
)

func AddMetric[T any](key string, f func() (T, error)) (T, error) {
	v, err := metricGaugeMap.Incr(key, func() (any, error) {
		ret, err := f()
		return ret, err
	})
	return v.(T), err
}

func GetMetrics() *GaugeMap {
	return metricGaugeMap
}
