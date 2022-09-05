// Command env is an example of confuguring profilers via environment variables.
package main

import (
	"flag"
	"log"

	"github.com/mmcloughlin/profile"
)

func main() {
	log.SetPrefix("example: ")
	log.SetFlags(0)

	// Configure flags.
	n := flag.Int("n", 1000000, "sum the integers 1 to `n`")
	flag.Parse()

	// Setup profiler.
	defer profile.Start(
		profile.AllProfiles,
		profile.ConfigEnvVar("PROFILE"),
	).Stop()

	// Sum 1 to n.
	sum := 0
	for i := 1; i <= *n; i++ {
		sum += i
	}
	log.Printf("sum: %d", sum)
}
