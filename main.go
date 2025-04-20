package main

import (
	"flag"
	"log"

	"github.com/piteego/enums/cmd"
)

func main() {
	var (
		// __type is the flag for the enum type name
		__type string
		// __output is the flag for the enum output file name
		//__output string
	)
	flag.StringVar(&__type, "type", "", "the name of the type")
	//flag.StringVar(&__output, "output", "", "the name of the output file")
	flag.Parse()
	if __type == "" {
		log.Fatalf("Type name is required")
	}
	if err := cmd.Generate(__type); err != nil {
		log.Fatalf("Failed to generate code: %v", err)
		return
	}
}
