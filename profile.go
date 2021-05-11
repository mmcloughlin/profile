// Package profile provides simple profiling for Go applications.
package profile

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
)

// Profile represents a profiling session.
type Profile struct {
	methods        []method
	log            func(string, ...interface{})
	noshutdownhook bool
	envvar         string

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

// ConfigEnvVar specifies an environment variable to configure profiles from.
func ConfigEnvVar(key string) func(*Profile) {
	return func(p *Profile) { p.envvar = key }
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

// config configures profiles based on a GODEBUG-like configuration string.
func (p *Profile) config(cfg string) {
	// Convert config string into equivalent command-line arguments and parse them.
	args := []string{}
	for _, arg := range strings.Split(cfg, ",") {
		args = append(args, "-"+arg)
	}

	// Register flags on a custom flagset. Register custom usage function that
	// will output flags in a format closer to the expected format of the
	// configuration string.
	f := flag.NewFlagSet("", flag.ExitOnError)
	p.SetFlags(f)

	f.Usage = func() {
		f.VisitAll(func(opt *flag.Flag) {
			value, usage := flag.UnquoteUsage(opt)
			fmt.Fprintf(f.Output(), "%s=%s\n\t%s\n", opt.Name, value, usage)
		})
	}

	// Parse. Discard error because ExitOnError ensures it's handled internally.
	_ = f.Parse(args)
}

// Start profiling.
func (p *Profile) Start() *Profile {
	// Set defaults.
	p.setdefaults()

	// Optionally configure via environment variable.
	if p.envvar != "" {
		p.config(os.Getenv(p.envvar))
	}

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
