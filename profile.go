// Package profile provides simple profiling for Go applications.
package profile

import (
	"flag"
	"io/ioutil"
	"log"
)

// Profile represents a profiling session.
type Profile struct {
	methods []method
	log     func(string, ...interface{})

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

// TODO: func NoShutdownHook(p *Profile)
// TODO: func ProfilePath(path string) func(*Profile)

// WithLogger configures informational messages to be logged to the given
// logger. Defaults to the standard library global logger.
func WithLogger(l *log.Logger) func(p *Profile) {
	return func(p *Profile) { p.log = l.Printf }
}

// Quiet suppresses logging.
func Quiet(p *Profile) {
	p.Configure(WithLogger(log.New(ioutil.Discard, "", 0)))
}

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
