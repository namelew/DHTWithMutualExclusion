package main

import "github.com/namelew/DHTWithMutualExclusion/internal/coordinator"

func main() {
	cd := coordinator.Build()
	cd.Handler()
}
