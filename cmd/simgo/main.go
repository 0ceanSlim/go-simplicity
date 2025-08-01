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
	help   = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *input == "" {
		fmt.Fprintf(os.Stderr, "Error: Input file is required\n\n")
		printUsage()
		os.Exit(1)
	}

	// Check if input file exists
	if _, err := os.Stat(*input); os.IsNotExist(err) {
		log.Fatalf("Input file does not exist: %s", *input)
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
		if *debug {
			fmt.Printf("Successfully compiled %s to %s\n", *input, *output)
		}
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s -input <go-file> [options]\n", os.Args[0])
	flag.PrintDefaults()
}

func printHelp() {
	fmt.Printf("go-simplicity - Go to Simplicity transpiler\n\n")
	fmt.Printf("USAGE:\n")
	fmt.Printf("    %s -input <go-file> [options]\n\n", os.Args[0])
	fmt.Printf("OPTIONS:\n")
	fmt.Printf("    -input string\n")
	fmt.Printf("        Input Go source file (required)\n")
	fmt.Printf("    -output string\n")
	fmt.Printf("        Output SimplicityHL file (default: stdout)\n")
	fmt.Printf("    -target string\n")
	fmt.Printf("        Target format: simplicityhl, simplicity (default: simplicityhl)\n")
	fmt.Printf("    -debug\n")
	fmt.Printf("        Enable debug output\n")
	fmt.Printf("    -help\n")
	fmt.Printf("        Show this help message\n\n")
	fmt.Printf("EXAMPLES:\n")
	fmt.Printf("    # Compile to stdout\n")
	fmt.Printf("    %s -input examples/basic_swap.go\n\n", os.Args[0])
	fmt.Printf("    # Compile to file\n")
	fmt.Printf("    %s -input examples/basic_swap.go -output basic_swap.shl\n\n", os.Args[0])
	fmt.Printf("    # Enable debug output\n")
	fmt.Printf("    %s -input examples/basic_swap.go -debug\n\n", os.Args[0])
}
