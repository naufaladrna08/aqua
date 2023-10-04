package main

import (
	"log"

	aqua "github.com/naufaladrna08/aqua/pkg/cmd/aqua"
)

func main() {
	err := aqua.RunServer()

	if err != nil {
		log.Fatalf("Error while registering gRPC service: %s\n", err.Error())
	}
}
