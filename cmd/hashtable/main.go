package main

import "github.com/namelew/DHTWithMutualExclusion/internal/server"

func main() {
	pid := server.New(0)
	pid.Build()
}
