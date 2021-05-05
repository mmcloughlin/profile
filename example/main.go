package main

import (
	"flag"
	"fmt"

	"github.com/mmcloughlin/profile"
)

func main() {
	n := flag.Int("n", 1000000, "sum the integers 1 to `n`")

	p := profile.New(
		profile.CPUProfile,
		profile.MemProfile,
		profile.BlockProfile,
		profile.MutexProfile,
		profile.GoroutineProfile,
		profile.ThreadcreationProfile,
		profile.TraceProfile,
	)

	p.SetFlags(flag.CommandLine)

	flag.Parse()

	// Start profilers.
	p.Start()

	// Sum 1 to n.
	sum := 0
	for i := 1; i <= *n; i++ {
		sum += i
	}
	fmt.Println(sum)

	// Stop profilers.
	p.Stop()
}
