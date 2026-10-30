// Package loot implements vanilla Bedrock Edition loot tables. The tables are
// embedded from Mojang's bedrock-samples repository
// (behavior_pack/loot_tables) and evaluated at runtime, so that loot follows
// vanilla exactly and can be updated by replacing the JSON files.
//
// Only the parts of the loot table format used by the embedded tables are
// supported: Pools with (ranged) rolls, item, loot_table and empty entries with
// weights and qualities, and the set_count, set_data, set_damage and
// enchant_with_levels functions. Conditions are not evaluated.
package loot

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"math/rand/v2"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/worldupgrader/itemupgrader"
)

//go:embed loot_tables
var files embed.FS

// Paths of the loot tables embedded in this package.
const (
	// Fishing is the loot table used for fishing outside of jungles.
	Fishing = "loot_tables/gameplay/fishing.json"
	// JungleFishing is the loot table used for fishing in jungle biomes.
	JungleFishing = "loot_tables/gameplay/jungle_fishing.json"
)

// Table is a vanilla loot table. It consists of pools, each of which rolls
// a number of times to produce items.
type Table struct {
	Pools []Pool `json:"pools"`
}

// Pool is a single pool of a loot table. Every roll of the pool picks one of
// its entries, weighted by the weight and quality of the entries.
type Pool struct {
	Rolls   Range   `json:"rolls"`
	Entries []Entry `json:"entries"`
}

// Entry is an entry in a loot table pool.
type Entry struct {
	// Type is the type of the entry: "item", "loot_table" or "empty".
	Type string `json:"type"`
	// Name is the name of the item for item entries, or the path of the loot
	// table for loot_table entries.
	Name string `json:"name"`
	// Weight is the base weight of the entry. It defaults to 1.
	Weight *int `json:"weight"`
	// Quality changes the weight of the entry by Quality*Luck.
	Quality   int        `json:"quality"`
	Functions []Function `json:"functions"`
}

// Function is a function applied to the item of an entry, such as setting its
// count or enchanting it.
type Function struct {
	// Function is the name of the function, such as "set_count". An optional
	// minecraft: prefix is ignored.
	Function string `json:"function"`
	// Count is used by set_count.
	Count Range `json:"count"`
	// Data is used by set_data.
	Data Range `json:"data"`
	// Damage is used by set_damage. It is the fraction of the item's durability
	// that remains, in the range 0-1.
	Damage Range `json:"damage"`
	// Levels and Treasure are used by enchant_with_levels.
	Levels   Range `json:"levels"`
	Treasure bool  `json:"treasure"`
}

// name returns the name of the function without namespace.
func (f Function) name() string {
	return strings.TrimPrefix(f.Function, "minecraft:")
}

// Range is a number in a loot table that is either constant or a range of
// the form {"min": x, "max": y}.
type Range struct {
	Min, Max float64
}

// UnmarshalJSON ...
func (r *Range) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err == nil {
		r.Min, r.Max = f, f
		return nil
	}
	var m struct{ Min, Max float64 }
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("decode range %s: %w", b, err)
	}
	r.Min, r.Max = m.Min, m.Max
	return nil
}

// Float returns a random float between the minimum and maximum of the range.
func (r Range) Float(rnd *rand.Rand) float64 {
	return r.Min + rnd.Float64()*(r.Max-r.Min)
}

// Int returns a random integer between the minimum and maximum of the range,
// both inclusive.
func (r Range) Int(rnd *rand.Rand) int {
	lo, hi := int(math.Floor(r.Min)), int(math.Floor(r.Max))
	if hi <= lo {
		return lo
	}
	return lo + rnd.IntN(hi-lo+1)
}

// Context holds the parameters used to generate loot from a Table.
type Context struct {
	// Luck changes the weight of entries by their quality. It is increased by,
	// for example, the Luck of the Sea enchantment.
	Luck float64
	// Rand is the source of randomness used. If nil, a random source is used.
	Rand *rand.Rand
}

var tables = sync.OnceValue(func() map[string]*Table {
	m := map[string]*Table{}
	err := fs.WalkDir(files, "loot_tables", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := files.ReadFile(path)
		if err != nil {
			return err
		}
		t := &Table{}
		if err := json.Unmarshal(b, t); err != nil {
			return fmt.Errorf("decode loot table %v: %w", path, err)
		}
		m[path] = t
		return nil
	})
	if err != nil {
		panic(err)
	}
	return m
})

// Lookup looks up the embedded vanilla loot table with the path passed, for
// example loot.Fishing. False is returned if no such table exists.
func Lookup(path string) (*Table, bool) {
	t, ok := tables()[path]
	return t, ok
}

// Generate generates loot from the table using the Context passed.
func (t *Table) Generate(ctx Context) []item.Stack {
	if ctx.Rand == nil {
		ctx.Rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	var stacks []item.Stack
	for _, pool := range t.Pools {
		for range pool.Rolls.Int(ctx.Rand) {
			stacks = append(stacks, pool.roll(ctx)...)
		}
	}
	return stacks
}

// roll rolls the pool once, picking an entry and generating its loot.
func (p Pool) roll(ctx Context) []item.Stack {
	total := 0
	weights := make([]int, len(p.Entries))
	for i, e := range p.Entries {
		if !e.available() {
			// Entries of items that are not implemented are left out, so that
			// the relative weights of the other entries stay the same.
			continue
		}
		w := 1
		if e.Weight != nil {
			w = *e.Weight
		}
		weights[i] = max(int(math.Floor(float64(w)+float64(e.Quality)*ctx.Luck)), 0)
		total += weights[i]
	}
	if total == 0 {
		return nil
	}
	r := ctx.Rand.IntN(total)
	for i, e := range p.Entries {
		if r -= weights[i]; r < 0 {
			return e.generate(ctx)
		}
	}
	return nil
}

// available checks if the entry can produce loot.
func (e Entry) available() bool {
	switch e.Type {
	case "item":
		_, ok := e.item(0)
		return ok
	case "loot_table":
		_, ok := Lookup(e.Name)
		return ok
	case "empty":
		return true
	}
	return false
}

// item resolves the item of an item entry with the data value passed.
// Legacy item names and data values used by the loot tables, such as
// minecraft:fish or minecraft:dye with data 3, are upgraded first.
func (e Entry) item(data int16) (world.Item, bool) {
	if data == 0 {
		for _, f := range e.Functions {
			if f.name() == "set_data" {
				// Use the lowest data value to check if the item exists.
				data = int16(f.Data.Min)
			}
		}
	}
	upgraded := itemupgrader.Upgrade(itemupgrader.ItemMeta{Name: e.Name, Meta: data})
	return world.ItemByName(upgraded.Name, upgraded.Meta)
}

// generate generates the loot of the entry.
func (e Entry) generate(ctx Context) []item.Stack {
	switch e.Type {
	case "loot_table":
		t, _ := Lookup(e.Name)
		return t.Generate(ctx)
	case "item":
		var data int16
		for _, f := range e.Functions {
			if f.name() == "set_data" {
				data = int16(f.Data.Int(ctx.Rand))
			}
		}
		it, ok := e.item(data)
		if !ok {
			return nil
		}
		s := item.NewStack(it, 1)
		for _, f := range e.Functions {
			s = f.apply(s, ctx.Rand)
		}
		if s.Empty() {
			return nil
		}
		return []item.Stack{s}
	}
	return nil
}

// apply applies the function to the item.Stack passed.
func (f Function) apply(s item.Stack, rnd *rand.Rand) item.Stack {
	switch f.name() {
	case "set_count":
		return s.Grow(f.Count.Int(rnd) - s.Count())
	case "set_damage":
		if maxDurability := s.MaxDurability(); maxDurability > 0 {
			remaining := math.Min(math.Max(f.Damage.Float(rnd), 0), 1)
			return s.WithDurability(max(int(math.Floor(remaining*float64(maxDurability))), 1))
		}
	case "enchant_with_levels":
		return EnchantWithLevels(s, f.Levels.Int(rnd), f.Treasure, rnd)
	}
	return s
}
