package main

import (
	"github.com/piteego/enums/cmd/generate"
	"log"
)

func main() {
	if err := generate.Execute(); err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}
}
