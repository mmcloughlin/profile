package profile

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

type Method interface {
	Name() string
	SetFlags(f *flag.FlagSet)
	Enabled() bool
	Start() error
	Stop() error
}

// cpu
//
//     - configure:
//         // runtime.SetCPUProfileRate(hz) // actually overridden by StartCPUProfile
//     - start:
//         open file
//         pprof.StartCPUProfile(f)
//     - stop:
//         pprof.StopCPUProfile()
//         close file
//     - flags:
//         -cpuprofile cpu.out
//

type CPU struct {
	filename string
	f        io.WriteCloser
}

func (CPU) Name() string { return "cpu" }

func (c *CPU) SetFlags(f *flag.FlagSet) {
	// Reference: https://github.com/golang/go/blob/303b194c6daf319f88e56d8ece56d924044f65a8/src/testing/testing.go#L292
	//
	//		cpuProfile = flag.String("test.cpuprofile", "", "write a cpu profile to `file`")
	//
	f.StringVar(&c.filename, "cpuprofile", "", "write a cpu profile to `file`")
}

func (c *CPU) Enabled() bool { return c.filename != "" }

func (c *CPU) Start() error {
	// Open output file.
	f, err := os.Create(c.filename)
	if err != nil {
		return err
	}

	c.f = f

	// Start profile.
	if err := pprof.StartCPUProfile(f); err != nil {
		return err
	}

	return nil
}

func (c *CPU) Stop() error {
	pprof.StopCPUProfile()
	return c.f.Close()
}

// mem
//
//     - configure:
//         runtime.MemProfileRate
//     - start:
//         N/A
//     - stop:
//         runtime.GC()
//         pprof.Lookup("allocs").WriteTo(f, 0)
//         close file
//         restore runtime.MemProfileRate
//     - flags:
//         -memprofile mem.out
//         -memprofilerate n

type Mem struct {
	filename string
	rate     int

	prevrate int
}

func (Mem) Name() string { return "mem" }

func (m *Mem) SetFlags(f *flag.FlagSet) {
	// Reference: https://github.com/golang/go/blob/303b194c6daf319f88e56d8ece56d924044f65a8/src/testing/testing.go#L290-L291
	//
	//		memProfile = flag.String("test.memprofile", "", "write an allocation profile to `file`")
	//		memProfileRate = flag.Int("test.memprofilerate", 0, "set memory allocation profiling `rate` (see runtime.MemProfileRate)")
	//
	flag.StringVar(&m.filename, "memprofile", "", "write an allocation profile to `file`")
	flag.IntVar(&m.rate, "memprofilerate", 0, "set memory allocation profiling `rate` (see runtime.MemProfileRate)")
}

func (m *Mem) Enabled() bool { return m.filename != "" }

func (m *Mem) Start() error {
	m.prevrate = runtime.MemProfileRate
	if m.rate > 0 {
		runtime.MemProfileRate = m.rate
	}
	return nil
}

func (m *Mem) Stop() error {
	// Materialize all statistics.
	runtime.GC()

	// Write to file.
	err := writeprofile("allocs", m.filename)

	// Restore profile rate.
	runtime.MemProfileRate = m.prevrate

	return err
}

// goroutine
//
//     - configure: N/A
//     - start: N/A
//     - stop:
//         open file
//         pprof.Lookup("goroutine").WriteTo
//         close file
//     - flags: N/A

// threadcreate
//
//     - configure: N/A
//     - start: N/A
//     - stop:
//         open file
//         pprof.Lookup("threadcreate").WriteTo
//         close file
//     - flags: N/A

// block
//
//     - configure:
//         runtime.SetBlockProfileRate
//     - start: N/A
//     - stop:
//         open file
//         pprof.Lookup("block").WriteTo
//         close file
//         reset SetBlockProfileRate
//     - flags:
//         -blockprofile block.out
//         -blockprofilerate n
//
// mutex
//
//     - configure:
//         runtime.SetMutexProfileFraction
//     - start: N/A
//     - stop:
//         open file
//         pprof.Lookup("mutex").WriteTo
//         close file
//         reset SetMutexProfileFraction
//     - flags:
//         -mutexprofile mutex.out
//         -mutexprofilefraction n
//
// trace
//
//     - configure: N/A
//     - start:
//         open file
//         trace.Start(w)
//     - stop:
//         trace.Stop()
//     - flags:
//         -trace trace.out

func writeprofile(name, filename string) error {
	// Open file.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	// Lookup profile.
	p := pprof.Lookup(name)
	if p == nil {
		return fmt.Errorf("unknown profile %q", name)
	}

	// Write.
	if err := p.WriteTo(f, 0); err != nil {
		return err
	}

	return f.Close()
}
