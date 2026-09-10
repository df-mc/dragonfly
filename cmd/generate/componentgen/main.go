package componentgen

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/df-mc/dragonfly/cmd/generate/componentgen/generate"
	"github.com/df-mc/dragonfly/cmd/generate/componentgen/parse"
	"github.com/df-mc/dragonfly/cmd/generate/generator"
)

// Generator implements the generate.Generator interface for component generation.
type Generator struct {
	outputDir  string
	vanillaNBT string
}

func init() {
	generator.Register(&Generator{})
}

func (g *Generator) Name() string {
	return "componentgen"
}

func (g *Generator) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&g.outputDir, "output", "", "Output directory for generated files")
	fs.StringVar(&g.vanillaNBT, "vanilla-nbt", "", "Path to vanilla_items.nbt")
}

func (g *Generator) Generate(ctx context.Context) error {
	if g.outputDir == "" {
		g.outputDir = filepath.Join("server", "item", "component")
	}
	if g.vanillaNBT == "" {
		g.vanillaNBT = filepath.Join("server", "world", "vanilla_items.nbt")
	}

	// Parse vanilla items NBT
	items, err := parse.VanillaItems(g.vanillaNBT)
	if err != nil {
		return fmt.Errorf("failed to parse vanilla items: %w", err)
	}

	// Extract component schemas
	schemas := parse.ExtractComponentSchemas(items)

	// Generate files
	gen := generate.NewGenerator(g.outputDir, schemas)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Println("Component generation complete")
	return nil
}
