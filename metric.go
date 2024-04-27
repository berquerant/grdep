package grdep

import (
	"strings"
	"sync"
	"time"
)

type GaugeMapIface interface {
	Walk(f func(key string, value *GaugeMapElement))
	Close()
	Incr(key string, f func() (any, error)) (any, error)
}

var (
	_ GaugeMapIface = &GaugeMap{}
	_ GaugeMapIface = &NullGaugeMap{}
)

type NullGaugeMap struct{}

func NewNullGaugeMap() *NullGaugeMap {
	return &NullGaugeMap{}
}

func (*NullGaugeMap) Walk(_ func(key string, value *GaugeMapElement)) {}
func (*NullGaugeMap) Close()                                          {}
func (*NullGaugeMap) Incr(_ string, f func() (any, error)) (any, error) {
	return f()
}

type GaugeMap struct {
	d     map[string]*GaugeMapElement
	addC  chan *addGaugeMapRequest
	doneC chan struct{}
}

type addGaugeMapRequest struct {
	key      string
	count    uint64
	duration time.Duration
}

type GaugeMapElement struct {
	Count    uint64
	Duration time.Duration
}

func (g *GaugeMap) Walk(f func(key string, value *GaugeMapElement)) {
	for k, v := range g.d {
		f(k, v)
	}
}

func (g *GaugeMap) Close() {
	close(g.addC)
	<-g.doneC
}

func (g *GaugeMap) addWorker() {
	for r := range g.addC {
		g.add(r.key, r.count, r.duration)
	}
	close(g.doneC)
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

	send := func(suffix string) {
		g.addC <- &addGaugeMapRequest{
			key:      key + "-" + suffix,
			count:    1,
			duration: elapsed,
		}
	}
	send("call")
	if err != nil {
		send("error")
	} else {
		send("success")
	}

	return ret, err
}

func newGaugeMap() *GaugeMap {
	g := &GaugeMap{
		d:     map[string]*GaugeMapElement{},
		addC:  make(chan *addGaugeMapRequest, 1000),
		doneC: make(chan struct{}),
	}
	go g.addWorker()
	return g
}

var (
	metricGaugeMap GaugeMapIface = NewNullGaugeMap()
	initMetrics                  = sync.OnceFunc(func() {
		metricGaugeMap = newGaugeMap()
	})
)

func InitMetrics() {
	initMetrics()
}

func AddMetric[T any](key string, f func() (T, error)) (T, error) {
	v, err := metricGaugeMap.Incr(key, func() (any, error) {
		ret, err := f()
		return ret, err
	})
	return v.(T), err
}

func GetMetrics() GaugeMapIface {
	return metricGaugeMap
}
