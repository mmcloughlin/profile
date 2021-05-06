package profile

import (
	"flag"
	"io/ioutil"
	"log"
)

type Profile struct {
	methods []method
	log     *log.Logger

	running []method
}

func New(options ...func(*Profile)) *Profile {
	p := &Profile{}
	p.Configure(options...)
	return p
}

func Start(options ...func(*Profile)) *Profile {
	return New(options...).Start()
}

func (p *Profile) Configure(options ...func(*Profile)) {
	for _, option := range options {
		option(p)
	}
}

// TODO: func NoShutdownHook(p *Profile)
// TODO: func ProfilePath(path string) func(*Profile)

func WithLogger(l *log.Logger) func(p *Profile) {
	return func(p *Profile) { p.log = l }
}

func Quiet(p *Profile) {
	p.Configure(WithLogger(log.New(ioutil.Discard, "", 0)))
}

func (p *Profile) addmethod(m method) {
	p.methods = append(p.methods, m)
}

func (p *Profile) SetFlags(f *flag.FlagSet) {
	for _, m := range p.methods {
		m.SetFlags(f)
	}
}

func (p *Profile) Start() *Profile {
	// Set defaults.
	if len(p.methods) == 0 {
		p.Configure(CPUProfile)
	}

	if p.log == nil {
		p.Configure(WithLogger(log.Default()))
	}

	// Start methods.
	for _, m := range p.methods {
		if !m.Enabled() {
			continue
		}

		if err := m.Start(); err != nil {
			p.log.Printf("%s profile: error starting: %v", m.Name(), err)
			continue
		}

		p.log.Printf("%s profile: started", m.Name())
		p.running = append(p.running, m)
	}

	return p
}

func (p *Profile) Stop() {
	for _, m := range p.running {
		if err := m.Stop(); err != nil {
			p.log.Printf("%s profile: error stopping: %v", m.Name(), err)
		} else {
			p.log.Printf("%s profile: stopped", m.Name())
		}
	}

	p.running = nil
}
