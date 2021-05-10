# profile

Simple profiling for Go

## Usage

The following example shows how to configure `profile` via flags with multiple
available profile types.

[embedmd]:# (internal/example/flags/main.go)
```go
package main

import (
	"flag"
	"log"

	"github.com/mmcloughlin/profile"
)

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
Usage of flags:
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
flags -n 1000000000 -cpuprofile cpu.out -memprofile mem.out
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
