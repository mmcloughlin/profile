// Command multi is an example of enabling multiple profiles at once.
package main

import "github.com/mmcloughlin/profile"

func main() {
	defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()
	// ...
}
