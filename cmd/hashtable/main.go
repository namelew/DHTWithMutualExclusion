package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/namelew/DHTWithMutualExclusion/internal/server"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Panic(err.Error())
	}
	pid := server.New(0)
	pid.Build()
}
