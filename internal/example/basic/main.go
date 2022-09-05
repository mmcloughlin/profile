// Command basic is an example of basic use of the profile package.
package main

import "github.com/mmcloughlin/profile"

func main() {
	defer profile.Start().Stop()
	// ...
}
