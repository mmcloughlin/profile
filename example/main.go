package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mmcloughlin/profile"
)

func main() {
	n := flag.Int("n", 1000000, "sum the integers 1 to `n`")

	r := &profile.Runner{
		Methods: []profile.Method{
			&profile.CPU{},
			&profile.Mem{},
		},
		Logger: log.Default(),
	}

	r.SetFlags(flag.CommandLine)
	flag.Parse()

	// Start profilers.
	r.Start()

	// Sum 1 to n.
	sum := 0
	for i := 1; i <= *n; i++ {
		sum += i
	}
	fmt.Println(sum)

	// Stop profilers.
	r.Stop()
}
