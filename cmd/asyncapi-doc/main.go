package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fedanant/asyncapi-doc/internal/asyncapi"
)

// Build information set via ldflags.
var (
	Version   = "v0.0.1"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var srcFolder = "."
var outFile = "./api.yaml"

func init() {
	flag.StringVar(&srcFolder, "f", srcFolder, "folder project")
	flag.StringVar(&outFile, "o", outFile, "output file")
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: command is required\n\n")
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generate()
	case "version", "--version", "-v":
		fmt.Printf("asyncapi-doc version %s\n", Version)
		fmt.Printf("  Build time: %s\n", BuildTime)
		fmt.Printf("  Git commit: %s\n", GitCommit)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func generate() {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	output := fs.String("output", "./asyncapi.yaml", "output file for generated AsyncAPI specification")
	verbose := fs.Bool("verbose", false, "enable verbose output")

	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("Failed to parse flags: %v\n", err)
	}

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: source directory is required\n")
		fmt.Fprintf(os.Stderr, "Usage: asyncapi-doc generate [options] <source-directory>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	codeFolder := fs.Arg(0)

	if *verbose {
		fmt.Printf("Parsing source directory: %s\n", codeFolder)
		fmt.Printf("Output file: %s\n", *output)
	}

	yaml, err := asyncapi.ParseFolder(codeFolder, *verbose)
	if err != nil {
		log.Fatalf("Failed to parse folder: %v\n", err)
	}

	if *verbose {
		fmt.Printf("Writing output to: %s\n", *output)
	}

	if err := os.WriteFile(*output, yaml, 0o600); err != nil {
		log.Fatalf("Failed to write output file: %v\n", err)
	}

	fmt.Println("✓ AsyncAPI specification generated successfully!")
}

func printUsage() {
	fmt.Printf(`asyncapi-doc - AsyncAPI Documentation Generator CLI Tool (v%s)

Usage:
  asyncapi-doc <command> [options] [arguments]

Available Commands:
  generate    Generate AsyncAPI specification from Go code
  version     Print version information
  help        Show this help message

Examples:
  asyncapi-doc generate -output ./asyncapi.yaml ./example/nats

Use "asyncapi-doc <command> -h" for more information about a command.
`, Version)
}
