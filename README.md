# profile

[![Go Reference](https://pkg.go.dev/badge/github.com/mmcloughlin/profile.svg)](https://pkg.go.dev/github.com/mmcloughlin/profile)

Simple profiling for Go.

* Easy management of Go's built-in
  [profiling](https://golang.org/pkg/runtime/pprof) and
  [tracing](https://golang.org/pkg/runtime/trace)
* Based on the widely-used [`pkg/profile`](https://github.com/pkg/profile):
  mostly-compatible API
* Supports generating multiple profiles at once
* Configurable with [idiomatic flags](#flags): `-cpuprofile`, `-memprofile`, ...
  just like `go test`
* Configurable by [environment variable](#environment): key-value interface like
  `GODEBUG`

## Install

```
go get github.com/mmcloughlin/profile
```

## Usage

Enabling profiling in your application is as simple as one line at the top of
your main function.

[embedmd]:# (internal/example/basic/main.go go /import/ /^}/)
```go
import "github.com/mmcloughlin/profile"

func main() {
	defer profile.Start().Stop()
	// ...
}
```

This will write a CPU profile to the current directory. Generate multiple
profiles by passing options to the `Start` function.

[embedmd]:# (internal/example/multi/main.go go /defer.*/)
```go
defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()
```

Profiles can also be configured by the user via [flags](#flags) or [environment
variable](#environment), as demonstrated in the examples below.

### Flags

The following example shows how to configure `profile` via flags with multiple
available profile types.

[embedmd]:# (internal/example/flags/main.go /func main/ /^}/)
```go
func main() {
	log.SetPrefix("example: ")
	log.SetFlags(0)

	// Setup profiler.
	p := profile.New(
		profile.CPUProfile,
		profile.MemProfile,
		profile.TraceProfile,
	)

	// Configure flags.
	n := flag.Int("n", 1000000, "sum the integers 1 to `n`")
	p.SetFlags(flag.CommandLine)
	flag.Parse()

	// Start profiler.
	defer p.Start().Stop()

	// Sum 1 to n.
	sum := 0
	for i := 1; i <= *n; i++ {
		sum += i
	}
	log.Printf("sum: %d", sum)
}
```

See the registered flags:

[embedmd]:# (internal/example/flags/help.err)
```err
Usage of example:
  -cpuprofile file
    	write a cpu profile to file
  -memprofile file
    	write an allocation profile to file
  -memprofilerate rate
    	set memory allocation profiling rate (see runtime.MemProfileRate)
  -n n
    	sum the integers 1 to n (default 1000000)
  -trace file
    	write an execution trace to file
```

Profile the application with the following flags:

[embedmd]:# (internal/example/flags/run.sh sh /.*cpuprofile.*/)
```sh
example -n 1000000000 -cpuprofile cpu.out -memprofile mem.out
```

We'll see additional logging in the output, as well as the profiles `cpu.out`
and `mem.out` written on exit.

[embedmd]:# (internal/example/flags/run.err)
```err
example: cpu profile: started
example: mem profile: started
example: sum: 500000000500000000
example: cpu profile: stopped
example: mem profile: stopped
```

### Environment

For a user-facing tool you may not want to expose profiling options via flags.
The `profile` package also offers configuration by environment variable, similar
to the `GODEBUG` option offered by the Go runtime.

[embedmd]:# (internal/example/env/main.go go /.*Setup.*/ /.*Stop.*/)
```go
	// Setup profiler.
	defer profile.Start(
		profile.AllProfiles,
		profile.ConfigEnvVar("PROFILE"),
	).Stop()
```

Now you can enable profiling with an environment variable, as follows:

[embedmd]:# (internal/example/env/run.sh sh /.*cpuprofile.*/)
```sh
PROFILE=cpuprofile=cpu.out,memprofile=mem.out example -n 1000000000
```

The output will be just the same as for the previous flags example. Set the
environment variable to `help` to get help on available options:

[embedmd]:# (internal/example/env/help.sh)
```sh
PROFILE=help example
```

In this case you'll see:

[embedmd]:# (internal/example/env/help.err)
```err
blockprofile=file
	write a goroutine blocking profile to file
blockprofilerate=rate
	set blocking profile rate (see runtime.SetBlockProfileRate)
cpuprofile=file
	write a cpu profile to file
goroutineprofile=file
	write a running goroutine profile to file
memprofile=file
	write an allocation profile to file
memprofilerate=rate
	set memory allocation profiling rate (see runtime.MemProfileRate)
mutexprofile=string
	write a mutex contention profile to the named file after execution
mutexprofilefraction=int
	if >= 0, calls runtime.SetMutexProfileFraction()
threadcreateprofile=file
	write a thread creation profile to file
trace=file
	write an execution trace to file
```

## Thanks

Thank you to [Dave Cheney](https://dave.cheney.net/) and
[contributors](https://github.com/pkg/profile/graphs/contributors) for the
excellent [`pkg/profile`](https://github.com/pkg/profile) package, which
provided the inspiration and basis for this work.

## License

`profile` is available under the [BSD 3-Clause License](LICENSE). The license
retains the copyright notice from
[`pkg/profile`](https://github.com/pkg/profile).
