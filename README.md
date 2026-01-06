# AsyncAPI Generator

A CLI tool for working with AsyncAPI specifications. Generate code, validate specifications, and manage AsyncAPI-based projects.

## Features

- Generate code from AsyncAPI specifications
- Validate AsyncAPI specification files
- Initialize new AsyncAPI projects
- Built with Go standard library (no external dependencies for core functionality)

## Installation

### Build from source

```bash
make build
```

Or manually:

```bash
go build -o asyncapi-generator ./cmd/asyncapi-generator
```

### Install

```bash
make install
```

Or manually:

```bash
go install ./cmd/asyncapi-generator
```

## Usage

### Generate Code

Generate code from an AsyncAPI specification:

```bash
asyncapi-generator generate -output ./generated -template go asyncapi.yaml
```

Options:
- `-output` - Output directory for generated files (default: ./output)
- `-template` - Template to use for generation (default: default)
- `-verbose` - Enable verbose output

### Validate Specification

Validate an AsyncAPI specification file:

```bash
asyncapi-generator validate asyncapi.yaml
```

Options:
- `-verbose` - Enable verbose output

### Initialize Project

Create a new AsyncAPI project with a sample specification:

```bash
asyncapi-generator init my-project
```

This will create:
- A new directory with the project name
- A sample `asyncapi.yaml` file
- A basic README

### Version

Print version information:

```bash
asyncapi-generator version
```

### Help

Show usage information:

```bash
asyncapi-generator help
```

## Project Structure

```
.
├── cmd/
│   └── asyncapi-generator/    # Main application entry point
│       └── main.go
├── internal/
│   ├── generator/              # Code generation logic
│   │   └── generator.go
│   └── config/                 # Configuration management
│       └── config.go
├── pkg/                        # Public packages (if any)
├── go.mod                      # Go module definition
├── Makefile                    # Build automation
└── README.md                   # This file
```

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Clean Build Artifacts

```bash
make clean
```

### Format Code

```bash
make fmt
```

### Run Linter

```bash
make lint
```

## Examples

### Basic Workflow

1. Initialize a new project:
```bash
asyncapi-generator init my-api
cd my-api
```

2. Edit the `asyncapi.yaml` file to define your API

3. Validate your specification:
```bash
asyncapi-generator validate asyncapi.yaml
```

4. Generate code:
```bash
asyncapi-generator generate asyncapi.yaml
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
