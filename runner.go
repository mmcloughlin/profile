package profile

import (
	"flag"
	"log"
)

type Runner struct {
	Methods []Method
	Logger  *log.Logger

	running []Method
}

func (r *Runner) SetFlags(f *flag.FlagSet) {
	for _, m := range r.Methods {
		m.SetFlags(f)
	}
}

func (r *Runner) Start() {
	for _, m := range r.Methods {
		if !m.Enabled() {
			continue
		}

		if err := m.Start(); err != nil {
			r.Logger.Printf("%s profile: error starting: %v", m.Name(), err)
			continue
		}

		r.Logger.Printf("%s profile: started", m.Name())
		r.running = append(r.running, m)
	}
}

func (r *Runner) Stop() {
	for _, m := range r.running {
		if err := m.Stop(); err != nil {
			r.Logger.Printf("%s profile: error stopping: %v", m.Name(), err)
		} else {
			r.Logger.Printf("%s profile: stopped", m.Name())
		}
	}

	r.running = nil
}
