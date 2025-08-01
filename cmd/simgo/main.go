package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
)

var (
	input  = flag.String("input", "", "Input Go source file")
	output = flag.String("output", "", "Output SimplicityHL file (default: stdout)")
	target = flag.String("target", "simplicityhl", "Target format: simplicityhl, simplicity")
	debug  = flag.Bool("debug", false, "Enable debug output")
)

func main() {
	flag.Parse()

	if *input == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -input <go-file> [-output <output-file>] [-target <target>]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read input file
	source, err := os.ReadFile(*input)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// Create compiler instance
	c := compiler.New(compiler.Config{
		Target: *target,
		Debug:  *debug,
	})

	// Compile Go source to target format
	result, err := c.Compile(string(source), *input)
	if err != nil {
		log.Fatalf("Compilation failed: %v", err)
	}

	// Write output
	if *output == "" {
		fmt.Print(result)
	} else {
		err := os.WriteFile(*output, []byte(result), 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Successfully compiled %s to %s\n", *input, *output)
	}
}
