package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fedanant/asyncapi-doc/internal/asyncapi"
)

const version = "0.1.0"

var srcFolder = "."
var outFile = "./api.yaml"

func init() {
	flag.StringVar(&srcFolder, "f", srcFolder, "folder project")
	flag.StringVar(&outFile, "o", outFile, "output file")
}

func main() {
	flag.Parse()

	command := os.Args[1]

	switch command {
	case "generate":
		generate()
	case "version":
		fmt.Printf("asyncapi-doc version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func generate() {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	output := fs.String("output", "./output", "output directory for generated files")
	template := fs.String("template", "default", "template to use for generation")
	verbose := fs.Bool("verbose", false, "verbose output")

	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: spec file is required\n")
		fmt.Fprintf(os.Stderr, "Usage: asyncapi-doc generate [options] <spec-file>\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	codeFolder := fs.Arg(0)

	if *verbose {
		fmt.Printf("Generating from specification: %s\n", codeFolder)
		fmt.Printf("Output directory: %s\n", *output)
		fmt.Printf("Template: %s\n", *template)
	}

	yaml, err := asyncapi.ParseFolder(codeFolder)
	if err != nil {
		log.Fatalln(err)
	}

	os.WriteFile(*output, yaml, 0o600)

	fmt.Println("✓ Code generation completed successfully!")
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
`, version)
}
