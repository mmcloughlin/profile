package main

import "github.com/mmcloughlin/profile"

func main() {
	defer profile.Start(profile.CPUProfile, profile.MemProfile).Stop()
	// ...
}
