package grdep

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

type Profiler struct {
	path string
}

func NewProfiler(path string) *Profiler {
	return &Profiler{
		path: path,
	}
}

func (p Profiler) newFile(name string) *os.File {
	f, err := os.Create(filepath.Join(p.path, name))
	if err != nil {
		p.fatal("profiler: failed to create profile", "name", name, "err", err)
	}
	return f
}

func (Profiler) fatal(msg string, args ...any) {
	L().Error(msg, args...)
	os.Exit(1)
}

// Start profiling.
//
// name: cpu, goroutine, threadcreate, heap, block, mutex
// See https://pkg.go.dev/runtime/pprof#Profile
func (p Profiler) Start(name string) func() {
	L().Info("profiler: start", "name", name)
	var (
		f        = p.newFile(name)
		stopInfo = func() {
			L().Info("profiler: stop", "name", name, "profile", f.Name())
		}
		write = func() {
			_ = pprof.Lookup(name).WriteTo(f, 0)
			f.Close()
			stopInfo()
		}
	)

	switch name {
	case "cpu":
		_ = pprof.StartCPUProfile(f)
		return func() {
			pprof.StopCPUProfile()
			f.Close()
			stopInfo()
		}
	case "goroutine", "threadcreate":
		return write
	case "heap":
		oldRate := runtime.MemProfileRate
		runtime.MemProfileRate = 1 << 12
		return func() {
			write()
			runtime.MemProfileRate = oldRate
		}
	case "block":
		runtime.SetBlockProfileRate(1)
		return func() {
			write()
			runtime.SetBlockProfileRate(0)
		}
	case "mutex":
		runtime.SetMutexProfileFraction(1)
		return func() {
			write()
			runtime.SetMutexProfileFraction(0)
		}

	default:
		f.Close()
		p.fatal("profiler: unknown profile", "name", name)
		return func() {}
	}
}
