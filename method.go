package profile

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

// AllProfiles enables all profiling types. Running all profiles at once is
// generally not a good idea, so it's recommended that this option is combined
// with some configuration mechanism, via flags or otherwise.
func AllProfiles(p *Profile) {
	p.Configure(
		CPUProfile,
		MemProfile,
		GoroutineProfile,
		ThreadcreationProfile,
		BlockProfile,
		MutexProfile,
		TraceProfile,
	)
}

type method interface {
	Name() string
	SetFlags(f *flag.FlagSet)
	Enabled() bool
	Start() error
	Stop() error
}

// CPUProfile enables cpu profiling.
func CPUProfile(p *Profile) {
	p.addmethod(&cpu{
		filename: "cpu.pprof",
	})
}

type cpu struct {
	filename string

	f io.WriteCloser
}

func (cpu) Name() string { return "cpu" }

func (c *cpu) SetFlags(f *flag.FlagSet) {
	// Reference: https://github.com/golang/go/blob/303b194c6daf319f88e56d8ece56d924044f65a8/src/testing/testing.go#L292
	//
	//		cpuProfile = flag.String("test.cpuprofile", "", "write a cpu profile to `file`")
	//
	f.StringVar(&c.filename, "cpuprofile", "", "write a cpu profile to `file`")
}

func (c *cpu) Enabled() bool { return c.filename != "" }

func (c *cpu) Start() error {
	// Open output file.
	f, err := os.Create(c.filename)
	if err != nil {
		return err
	}

	// Start profile.
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = f.Close() // best effort: ignore error since we already have one
		return err
	}

	c.f = f

	return nil
}

func (c *cpu) Stop() error {
	pprof.StopCPUProfile()
	return c.f.Close()
}

// MemProfile enables memory profiling.
func MemProfile(p *Profile) {
	p.addmethod(&mem{
		filename: "mem.pprof",
	})
}

type mem struct {
	filename string
	rate     int

	prevrate int
}

func (mem) Name() string { return "mem" }

func (m *mem) SetFlags(f *flag.FlagSet) {
	// Reference: https://github.com/golang/go/blob/303b194c6daf319f88e56d8ece56d924044f65a8/src/testing/testing.go#L290-L291
	//
	//		memProfile = flag.String("test.memprofile", "", "write an allocation profile to `file`")
	//		memProfileRate = flag.Int("test.memprofilerate", 0, "set memory allocation profiling `rate` (see runtime.MemProfileRate)")
	//
	f.StringVar(&m.filename, "memprofile", "", "write an allocation profile to `file`")
	f.IntVar(&m.rate, "memprofilerate", 0, "set memory allocation profiling `rate` (see runtime.MemProfileRate)")
}

func (m *mem) Enabled() bool { return m.filename != "" }

func (m *mem) Start() error {
	m.prevrate = runtime.MemProfileRate
	if m.rate > 0 {
		runtime.MemProfileRate = m.rate
	}
	return nil
}

func (m *mem) Stop() error {
	// Materialize all statistics.
	runtime.GC()

	// Write to file.
	err := writeprofile("allocs", m.filename)

	// Restore profile rate.
	runtime.MemProfileRate = m.prevrate

	return err
}

// GoroutineProfile enables goroutine profiling.
func GoroutineProfile(p *Profile) {
	p.addmethod(&lookup{
		name:     "goroutine",
		long:     "running goroutine",
		filename: "goroutine.pprof",
	})
}

// ThreadcreationProfile enables thread creation profiling.
func ThreadcreationProfile(p *Profile) {
	p.addmethod(&lookup{
		name:     "threadcreate",
		long:     "thread creation",
		filename: "threadcreate.pprof",
	})
}

type lookup struct {
	name string
	long string

	filename string
}

func (l *lookup) Name() string { return l.name }

func (l *lookup) SetFlags(f *flag.FlagSet) {
	f.StringVar(&l.filename, l.name+"profile", "", "write a "+l.long+" profile to `file`")
}

func (l *lookup) Enabled() bool { return l.filename != "" }

func (l *lookup) Start() error { return nil }

func (l *lookup) Stop() error {
	return writeprofile(l.name, l.filename)
}

// BlockProfile enables block (contention) profiling.
func BlockProfile(p *Profile) {
	p.addmethod(&block{
		filename: "block.pprof",
		rate:     1,
	})
}

type block struct {
	filename string
	rate     int
}

func (block) Name() string { return "block" }

func (b *block) SetFlags(f *flag.FlagSet) {
	// Reference: https://github.com/golang/go/blob/303b194c6daf319f88e56d8ece56d924044f65a8/src/testing/testing.go#L293-L294
	//
	//		blockProfile = flag.String("test.blockprofile", "", "write a goroutine blocking profile to `file`")
	//		blockProfileRate = flag.Int("test.blockprofilerate", 1, "set blocking profile `rate` (see runtime.SetBlockProfileRate)")
	//
	f.StringVar(&b.filename, "blockprofile", "", "write a goroutine blocking profile to `file`")
	f.IntVar(&b.rate, "blockprofilerate", 1, "set blocking profile `rate` (see runtime.SetBlockProfileRate)")
}

func (b *block) Enabled() bool { return b.filename != "" && b.rate > 0 }

func (b *block) Start() error {
	runtime.SetBlockProfileRate(b.rate)
	return nil
}

func (b *block) Stop() error {
	// Write to file.
	err := writeprofile("block", b.filename)

	// Disable block profiling.
	runtime.SetBlockProfileRate(0)

	return err
}

// MutexProfile enables mutex profiling.
func MutexProfile(p *Profile) {
	p.addmethod(&mutex{
		filename: "mutex.pprof",
		rate:     1,
	})
}

type mutex struct {
	filename string
	rate     int
}

func (mutex) Name() string { return "mutex" }

func (m *mutex) SetFlags(f *flag.FlagSet) {
	// Reference: https://github.com/golang/go/blob/303b194c6daf319f88e56d8ece56d924044f65a8/src/testing/testing.go#L295-L296
	//
	//		mutexProfile = flag.String("test.mutexprofile", "", "write a mutex contention profile to the named file after execution")
	//		mutexProfileFraction = flag.Int("test.mutexprofilefraction", 1, "if >= 0, calls runtime.SetMutexProfileFraction()")
	//
	f.StringVar(&m.filename, "mutexprofile", "", "write a mutex contention profile to the named file after execution")
	f.IntVar(&m.rate, "mutexprofilefraction", 1, "if >= 0, calls runtime.SetMutexProfileFraction()")
}

func (m *mutex) Enabled() bool { return m.filename != "" && m.rate > 0 }

func (m *mutex) Start() error {
	runtime.SetMutexProfileFraction(m.rate)
	return nil
}

func (m *mutex) Stop() error {
	// Write to file.
	err := writeprofile("mutex", m.filename)

	// Disable mutex profiling.
	runtime.SetMutexProfileFraction(0)

	return err
}

// TraceProfile enables execution tracing.
func TraceProfile(p *Profile) {
	p.addmethod(&tracer{
		filename: "trace.out",
	})
}

type tracer struct {
	filename string

	f io.WriteCloser
}

func (tracer) Name() string { return "trace" }

func (t *tracer) SetFlags(f *flag.FlagSet) {
	// Reference: https://github.com/golang/go/blob/303b194c6daf319f88e56d8ece56d924044f65a8/src/testing/testing.go#L298
	//
	//		traceFile = flag.String("test.trace", "", "write an execution trace to `file`")
	//
	f.StringVar(&t.filename, "trace", "", "write an execution trace to `file`")
}

func (t *tracer) Enabled() bool { return t.filename != "" }

func (t *tracer) Start() error {
	// Open output file.
	f, err := os.Create(t.filename)
	if err != nil {
		return err
	}

	// Start trace.
	if err := trace.Start(f); err != nil {
		_ = f.Close() // best effort: ignore error since we already have one
		return err
	}

	t.f = f

	return nil
}

func (t *tracer) Stop() error {
	trace.Stop()
	return t.f.Close()
}

func writeprofile(name, filename string) (err error) {
	// Lookup profile.
	p := pprof.Lookup(name)
	if p == nil {
		return fmt.Errorf("unknown profile %q", name)
	}

	// Open file.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if errc := f.Close(); err == nil && errc != nil {
			err = errc
		}
	}()

	// Write.
	return p.WriteTo(f, 0)
}
