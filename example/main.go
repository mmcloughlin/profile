package main

import (
	"flag"
	"log"

	"github.com/mmcloughlin/profile"
)

func main() {
	log.SetPrefix("example: ")
	log.SetFlags(0)

	n := flag.Int("n", 1000000, "sum the integers 1 to `n`")
	flag.Parse()

	// Start profiler.
	defer profile.Start().Stop()

	// Sum 1 to n.
	sum := 0
	for i := 1; i <= *n; i++ {
		sum += i
	}
	log.Printf("sum: %d", sum)
}
