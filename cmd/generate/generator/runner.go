package generator

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// Generator is the interface that all generators must implement.
type Generator interface {
	// Name returns the name of the generator.
	Name() string
	// Generate runs the generator.
	Generate(ctx context.Context) error
	// SetFlags sets the flags for the generator.
	SetFlags(*flag.FlagSet)
}

// Registry holds all registered generators.
var Registry = map[string]Generator{}

// Register registers a generator.
func Register(g Generator) {
	if _, ok := Registry[g.Name()]; ok {
		panic(fmt.Sprintf("generator %q already registered", g.Name()))
	}
	Registry[g.Name()] = g
}

// Run runs the specified generators. If names is empty, runs all generators.
func Run(ctx context.Context, names ...string) error {
	var generators []Generator
	if len(names) == 0 {
		for _, g := range Registry {
			generators = append(generators, g)
		}
	} else {
		for _, name := range names {
			g, ok := Registry[name]
			if !ok {
				return fmt.Errorf("unknown generator %q", name)
			}
			generators = append(generators, g)
		}
	}

	// Set up flags for each generator
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	for _, g := range generators {
		g.SetFlags(fs)
	}
	// Parse all remaining args after the generator name
	// names[0] is the generator name, so we need to skip it
	// We'll parse from the full os.Args but skip the first two elements (binary name + generator name)
	fs.Parse(os.Args[2:])

	for _, g := range generators {
		fmt.Printf("Running generator: %s\n", g.Name())
		if err := g.Generate(ctx); err != nil {
			return fmt.Errorf("generator %s failed: %w", g.Name(), err)
		}
	}
	return nil
}

// Main is the entry point for the generate command.
// It can be called from a go:generate directive.
func Main() {
	ctx := context.Background()

	// Get generator name from first arg
	args := os.Args[1:]
	if len(args) == 0 {
		// Run all generators
		if err := Run(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Generation failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run specific generator
	if err := Run(ctx, args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Generation failed: %v\n", err)
		os.Exit(1)
	}
}
