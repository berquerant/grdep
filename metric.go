package grdep

import (
	"fmt"
	"strings"
	"sync"
)

type GaugeMap struct {
	d   map[string]uint64
	mux sync.RWMutex
}

func (g *GaugeMap) Walk(f func(key string, value uint64)) {
	g.mux.RLock()
	defer g.mux.RUnlock()
	for k, v := range g.d {
		f(k, v)
	}
}

func (*GaugeMap) sanitizeKey(key string) string {
	return strings.ReplaceAll(key, " ", "_")
}

func (g *GaugeMap) Incr(key string, err error) {
	g.mux.Lock()
	defer g.mux.Unlock()

	key = g.sanitizeKey(key)
	g.d[fmt.Sprintf("%s-call", key)]++
	if err != nil {
		g.d[fmt.Sprintf("%s-error", key)]++
	} else {
		g.d[fmt.Sprintf("%s-success", key)]++
	}
}

func NewGaugeMap() *GaugeMap {
	return &GaugeMap{
		d: map[string]uint64{},
	}
}

var (
	metricGaugeMap = NewGaugeMap()
)

func AddMetric(key string, err error) {
	metricGaugeMap.Incr(key, err)
}

func GetMetrics() *GaugeMap {
	return metricGaugeMap
}
