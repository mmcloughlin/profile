// Package profile provides simple profiling for Go applications.
package profile

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

// Profile represents a profiling session.
type Profile struct {
	methods        []method
	log            func(string, ...interface{})
	noshutdownhook bool

	running []method
}

// New creates a new profiling session configured with the given options.
func New(options ...func(*Profile)) *Profile {
	p := &Profile{
		log: log.Printf,
	}
	p.Configure(options...)
	return p
}

// Start a new profiling session with the given options.
func Start(options ...func(*Profile)) *Profile {
	return New(options...).Start()
}

// Configure applies the given options to this profiling session.
func (p *Profile) Configure(options ...func(*Profile)) {
	for _, option := range options {
		option(p)
	}
}

// WithLogger configures informational messages to be logged to the given
// logger. Defaults to the standard library global logger.
func WithLogger(l *log.Logger) func(p *Profile) {
	return func(p *Profile) { p.log = l.Printf }
}

// Quiet suppresses logging.
func Quiet(p *Profile) {
	p.Configure(WithLogger(log.New(ioutil.Discard, "", 0)))
}

// NoShutdownHook controls whether the profiling session should shutdown on
// interrupt.  Programs with more sophisticated signal handling should use this
// option to disable the default shutdown handler, and ensure the profile Stop()
// method is called during shutdown.
func NoShutdownHook(p *Profile) { p.noshutdownhook = true }

func (p *Profile) addmethod(m method) {
	p.methods = append(p.methods, m)
}

func (p *Profile) setdefaults() {
	if len(p.methods) == 0 {
		p.Configure(CPUProfile)
	}
}

// SetFlags registers flags to configure this profiling session.  This should be
// called after all options have been applied.
func (p *Profile) SetFlags(f *flag.FlagSet) {
	p.setdefaults()
	for _, m := range p.methods {
		m.SetFlags(f)
	}
}

// Start profiling.
func (p *Profile) Start() *Profile {
	// Set defaults.
	p.setdefaults()

	// Start methods.
	for _, m := range p.methods {
		if !m.Enabled() {
			continue
		}

		if err := m.Start(); err != nil {
			p.log("%s profile: error starting: %v", m.Name(), err)
			continue
		}

		p.log("%s profile: started", m.Name())
		p.running = append(p.running, m)
	}

	// Shutdown hook.
	if !p.noshutdownhook {
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			s := <-c

			p.log("caught %v: stopping profiles", s)
			p.Stop()

			os.Exit(0)
		}()
	}

	return p
}

// Stop profiling.
func (p *Profile) Stop() {
	for _, m := range p.running {
		if err := m.Stop(); err != nil {
			p.log("%s profile: error stopping: %v", m.Name(), err)
		} else {
			p.log("%s profile: stopped", m.Name())
		}
	}

	p.running = nil
}
