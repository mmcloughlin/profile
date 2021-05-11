package main

import "log"

func init() {
	log.SetPrefix("example: ")
	log.SetFlags(0)
}
