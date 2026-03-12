// Command simgo compiles Go contract source files to SimplicityHL.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/0ceanslim/go-simplicity/pkg/compiler"
	"github.com/0ceanslim/go-simplicity/pkg/jets"
)

const version = "1.3.3"

var (
	input    = flag.String("input", "", "Input Go source file")
	output   = flag.String("output", "", "Output SimplicityHL file (default: stdout)")
	target   = flag.String("target", "simplicityhl", "Target format: simplicityhl, simplicity")
	debug    = flag.Bool("debug", false, "Enable debug output")
	help     = flag.Bool("help", false, "Show help message")
	listJets = flag.Bool("list-jets", false, "List all registered jets and exit")
	ver      = flag.Bool("version", false, "Print version and exit")
)

func main() {
	flag.Parse()

	if *ver {
		fmt.Printf("simgo version %s\n", version)
		return
	}

	if *listJets {
		printJets()
		return
	}

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
	fmt.Printf("    -list-jets\n")
	fmt.Printf("        List all registered jets and exit\n")
	fmt.Printf("    -version\n")
	fmt.Printf("        Print version and exit\n")
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

func printJets() {
	reg := jets.NewRegistry()
	all := reg.AllJets()

	// Collect and sort by Go name for deterministic output
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Printf("%-30s  %s\n", "Go name", "Simplicity name")
	fmt.Printf("%-30s  %s\n", "-------", "---------------")
	for _, name := range names {
		info := all[name]
		fmt.Printf("jet.%-26s  jet::%s\n", info.GoName, info.SimplicityName)
	}
	fmt.Printf("\n%d jets registered\n", len(names))
}
