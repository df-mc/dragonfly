package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// Bamboo is a non-solid plant block that can be placed on vegetation-supporting blocks.
type Bamboo struct {
	empty
	transparent
	Age      bool
	LeafSize int
	Thick    bool
}

var _ item.BoneMealAffected = Bamboo{}

// UseOnBlock places bamboo. When placed on soil it starts as a thin Bamboo Shoot (no leaves).
// When placed on top of an existing bamboo stalk it extends it and updates the whole stalk shape.
func (b Bamboo) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	below := pos.Side(cube.FaceDown)
	if _, ok := tx.Block(below).(Bamboo); !ok {
		if _, ok := tx.Block(below).(BambooSapling); !ok && !supportsVegetation(b, tx.Block(below)) {
			return false
		}
	}

	if _, ok := tx.Block(below).(Bamboo); ok {
		// Extending an existing stalk: new top block is Age=false (can grow).
		b.Age = false
		b.LeafSize = LargeLeaves
		b.Thick = true
		place(tx, pos, b, user, ctx)
		base := bambooBase(pos, tx)
		updateBambooStalk(base, tx)
	} else if _, ok := tx.Block(below).(BambooSapling); ok {
		// Placing on top of a sapling: convert sapling to bottom bamboo and extend stalk.
		b.Age = false
		b.LeafSize = LargeLeaves
		b.Thick = true
		place(tx, pos, b, user, ctx)
		// Convert the sapling below to a proper bamboo bottom block (aged).
		tx.SetBlock(below, Bamboo{Age: true, LeafSize: bambooNoLeaves, Thick: false}, nil)
		updateBambooStalk(below, tx)
	} else {
		// Planting a new shoot on the ground: use BambooSapling.
		sapling := BambooSapling{Age: false}
		place(tx, pos, sapling, user, ctx)
		// Seed the growth chain immediately after placement.
		tx.ScheduleBlockUpdate(pos, sapling, bambooGrowthDelay(nil))
		return placed(ctx)
	}
	// Seed the growth chain for the new bamboo top.
	if topPos, ok := bambooTop(pos, tx); ok {
		if topB, ok2 := tx.Block(topPos).(Bamboo); ok2 && !topB.Age {
			tx.ScheduleBlockUpdate(topPos, topB, bambooGrowthDelay(nil))
		}
	}
	return placed(ctx)
}

// BoneMeal grows a bamboo stalk by 1-2 blocks if there is enough room.
func (b Bamboo) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	top, ok := bambooTop(pos, tx)
	if !ok {
		return false
	}
	// The top block must have Age=false (growable) for bone meal to work.
	if topB, ok2 := tx.Block(top).(Bamboo); ok2 && topB.Age {
		return false
	}
	growth := rand.IntN(2) + 1
	applied := false
	for range growth {
		nextTop, ok := growBamboo(top, tx)
		if !ok {
			break
		}
		top = nextTop
		applied = true
	}
	if applied {
		// Re-seed the growth chain on the new top so natural growth resumes.
		if topB, ok2 := tx.Block(top).(Bamboo); ok2 && !topB.Age {
			tx.ScheduleBlockUpdate(top, topB, bambooGrowthDelay(nil))
		}
	}
	return applied
}

// NeighbourUpdateTick breaks the bamboo if it loses support.
func (b Bamboo) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !canSurviveBamboo(pos, tx) {
		breakBlock(b, pos, tx)
		tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
	}
}

// RandomTick handles survival checks and seeds the ScheduledTick growth chain
// for world-generated bamboo that was never explicitly placed by a player.
func (b Bamboo) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if !canSurviveBamboo(pos, tx) {
		breakBlock(b, pos, tx)
		return
	}
	// For world-generated bamboo: start the scheduled growth chain if the top
	// block has no pending tick yet. ScheduleBlockUpdate is a no-op when a
	// later tick is already queued, so this is safe to call unconditionally.
	if !b.Age {
		tx.ScheduleBlockUpdate(pos, b, bambooGrowthDelay(r))
	}
}

// ScheduledTick drives natural bamboo growth at vanilla-like speed (1-5 s per
// stage). Each firing grows one block and re-schedules the new top.
// Light < 9 (night / cave) causes a re-schedule without growing, matching the
// observed vanilla behaviour that bamboo does not grow at night.
func (b Bamboo) ScheduledTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if b.Age {
		return // stalk already at max height
	}
	above := pos.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return // blocked above — stop the chain
	}
	delay := bambooGrowthDelay(r)
	t := tx.World().Time() % 24000
	if t >= 13000 && t < 23000 {
		// Nighttime — wait and retry.
		tx.ScheduleBlockUpdate(pos, b, delay)
		return
	}
	newTop, ok := growBamboo(pos, tx)
	if !ok {
		return // reached per-stalk max height
	}
	if newTopB, ok2 := tx.Block(newTop).(Bamboo); ok2 && !newTopB.Age {
		tx.ScheduleBlockUpdate(newTop, newTopB, delay)
	}
}

// HasLiquidDrops ...
func (Bamboo) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo ...
func (Bamboo) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (b Bamboo) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       oneOf(b),
		BreakHandler: func(pos cube.Pos, tx *world.Tx, u item.User) {
			tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
			// When the top is broken, reset the new top to Age=false so it can
			// resume natural growth (mirrors PNX onBreak age reset logic).
			below := pos.Side(cube.FaceDown)
			if belowB, ok := tx.Block(below).(Bamboo); ok {
				base := bambooBase(below, tx)
				h := bambooHeightFromBase(base, tx)
				// Probabilistic stop mirrors PNX: always reset below height 11,
				// 75% chance to reset between 11-14, never reset at 15+.
				if h < 15 && (h < 11 || rand.IntN(4) != 0) {
					belowB.Age = false
					tx.SetBlock(below, belowB, nil)
				}
			}
		},
	}
}

// CompostChance ...
func (Bamboo) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (Bamboo) EncodeItem() (name string, meta int16) {
	return "minecraft:bamboo", 0
}

// EncodeBlock ...
func (b Bamboo) EncodeBlock() (string, map[string]any) {
	thickness := "thin"
	if b.Thick {
		thickness = "thick"
	}
	return "minecraft:bamboo", map[string]any{
		"age_bit":                boolByte(b.Age),
		"bamboo_leaf_size":       bambooLeafSizeString(b.LeafSize),
		"bamboo_stalk_thickness": thickness,
	}
}

// allBamboo returns all bamboo block states.
func allBamboo() (blocks []world.Block) {
	for _, age := range []bool{false, true} {
		for _, leafSize := range bambooLeafSizes() {
			for _, thick := range []bool{false, true} {
				blocks = append(blocks, Bamboo{Age: age, LeafSize: leafSize, Thick: thick})
			}
		}
	}
	return
}

const (
	bambooNoLeaves = iota
	SmallLeaves
	LargeLeaves
)

func bambooLeafSizes() []int {
	return []int{bambooNoLeaves, SmallLeaves, LargeLeaves}
}

func bambooLeafSizeString(size int) string {
	switch size {
	case SmallLeaves:
		return "small_leaves"
	case LargeLeaves:
		return "large_leaves"
	default:
		return "no_leaves"
	}
}

func canSurviveBamboo(pos cube.Pos, tx *world.Tx) bool {
	below := pos.Side(cube.FaceDown)
	if _, ok := tx.Block(below).(Bamboo); ok {
		return canSurviveBamboo(below, tx)
	}
	return supportsVegetation(Bamboo{}, tx.Block(below))
}

func bambooTop(pos cube.Pos, tx *world.Tx) (cube.Pos, bool) {
	if _, ok := tx.Block(pos).(Bamboo); !ok {
		return cube.Pos{}, false
	}
	for {
		next := pos.Side(cube.FaceUp)
		if _, ok := tx.Block(next).(Bamboo); !ok {
			return pos, true
		}
		pos = next
	}
}

func bambooBase(pos cube.Pos, tx *world.Tx) cube.Pos {
	for {
		next := pos.Side(cube.FaceDown)
		if _, ok := tx.Block(next).(Bamboo); !ok {
			return pos
		}
		pos = next
	}
}

func bambooHeightFromBase(base cube.Pos, tx *world.Tx) int {
	height := 1
	for curr := base.Side(cube.FaceUp); ; curr = curr.Side(cube.FaceUp) {
		if _, ok := tx.Block(curr).(Bamboo); !ok {
			return height
		}
		height++
	}
}

// growBamboo grows the bamboo stalk at top by one block.
// top must be the current top block (Age=false, air above).
// Uses a PNX-style local update: only touches the top 3–4 blocks,
// never the full stalk, so it stays fast even for tall bamboo.
func growBamboo(top cube.Pos, tx *world.Tx) (cube.Pos, bool) {
	above := top.Side(cube.FaceUp)
	if _, ok := tx.Block(above).(Air); !ok {
		return cube.Pos{}, false
	}
	base := bambooBase(top, tx)
	totalHeight := bambooHeightFromBase(base, tx)

	// Each bamboo stalk gets a deterministic max height in the 12-16 range.
	if totalHeight >= bambooMaxHeight(base) {
		return cube.Pos{}, false
	}

	topB, ok := tx.Block(top).(Bamboo)
	if !ok {
		return cube.Pos{}, false
	}

	// If the base is still a sapling, convert it to bamboo first.
	if _, ok2 := tx.Block(base).(BambooSapling); ok2 {
		tx.SetBlock(base, Bamboo{Age: true, LeafSize: bambooNoLeaves, Thick: false}, nil)
	}

	newHeight := totalHeight + 1
	becomesThick := newHeight >= 4

	switch {
	case topB.Thick:
		// Already thick: new top = thick + large_leaves.
		// Update top 3 blocks (mirrors PNX place() for thick parent).
		tx.SetBlock(above, Bamboo{Age: false, LeafSize: LargeLeaves, Thick: true}, nil)
		topB.Age = true
		topB.LeafSize = LargeLeaves
		topB.Thick = true
		tx.SetBlock(top, topB, nil)
		p1 := top.Side(cube.FaceDown)
		if b1, ok2 := tx.Block(p1).(Bamboo); ok2 {
			b1.Age = true
			b1.LeafSize = SmallLeaves
			b1.Thick = true
			tx.SetBlock(p1, b1, nil)
		}
		p2 := p1.Side(cube.FaceDown)
		if b2, ok2 := tx.Block(p2).(Bamboo); ok2 {
			b2.Age = true
			b2.LeafSize = bambooNoLeaves
			b2.Thick = true
			tx.SetBlock(p2, b2, nil)
		}
	case becomesThick:
		// Thin → thick transition at height 4.
		// New top: thick + large_leaves. Old top: thick + small_leaves.
		// All blocks below: thick + no_leaves (mirrors PNX setThick path).
		tx.SetBlock(above, Bamboo{Age: false, LeafSize: LargeLeaves, Thick: true}, nil)
		topB.Age = true
		topB.LeafSize = SmallLeaves
		topB.Thick = true
		tx.SetBlock(top, topB, nil)
		curr := top.Side(cube.FaceDown)
		for {
			bCurr, ok2 := tx.Block(curr).(Bamboo)
			if !ok2 {
				break
			}
			bCurr.Age = true
			bCurr.Thick = true
			bCurr.LeafSize = bambooNoLeaves
			tx.SetBlock(curr, bCurr, nil)
			curr = curr.Side(cube.FaceDown)
		}
	default:
		// Thin growth (height 1–3): new top = thin + small_leaves.
		tx.SetBlock(above, Bamboo{Age: false, LeafSize: SmallLeaves, Thick: false}, nil)
		topB.Age = true
		tx.SetBlock(top, topB, nil)
	}
	return above, true
}

// bambooGrowthDelay returns a random delay of 1-5 seconds for scheduled
// growth ticks, matching the vanilla-observed 1-5s per growth stage.
// r may be nil (uses the global rand source), which is valid in UseOnBlock.
func bambooGrowthDelay(r *rand.Rand) time.Duration {
	var n int
	if r != nil {
		n = r.IntN(81) // 0-80 ticks → 0-4s on top of the base 1s
	} else {
		n = rand.IntN(81)
	}
	return time.Duration(20+n) * time.Second / 20
}

// bambooMaxHeight returns a deterministic per-stalk maximum height in the
// range 12-16 inclusive, based on the bamboo base position.
func bambooMaxHeight(base cube.Pos) int {
	hash := uint32(base.X())*73428767 ^ uint32(base.Y())*912931 ^ uint32(base.Z())*43828943
	return 12 + int(hash%5)
}

// updateBambooStalk updates the entire bamboo stalk visuals and age bits.
// Called only on manual placement (UseOnBlock) – never on RandomTick for performance.
// Age=false is set on the top block only; all other blocks get Age=true.
// Height 1   = thin shoot, no leaves.
// Height 2-3 = thin stalk, top 2 blocks have large leaves.
// Height >=4 = thick stalk, top 2 blocks have large leaves, 3rd-from-top has small leaves.
func updateBambooStalk(base cube.Pos, tx *world.Tx) {
	height := bambooHeightFromBase(base, tx)
	if height == 0 {
		return
	}

	for i := 0; i < height; i++ {
		pos := base.Add(cube.Pos{0, i, 0})
		b, ok := tx.Block(pos).(Bamboo)
		if !ok {
			continue
		}

		// Only the topmost block can grow (age_bit=0 = Age=false in vanilla Bedrock).
		b.Age = (i != height-1)

		// Thickness: thin while short, thick once mature (height >= 4).
		b.Thick = height >= 4

		// Leaf distribution based on distance from the top.
		distFromTop := height - 1 - i
		switch height {
		case 1:
			b.LeafSize = bambooNoLeaves
		case 2:
			if distFromTop == 0 {
				b.LeafSize = LargeLeaves
			} else {
				b.LeafSize = bambooNoLeaves
			}
		case 3, 4:
			if distFromTop <= 1 {
				b.LeafSize = LargeLeaves
			} else {
				b.LeafSize = bambooNoLeaves
			}
		default: // height >= 5
			switch {
			case distFromTop <= 1:
				b.LeafSize = LargeLeaves
			case distFromTop == 2:
				b.LeafSize = SmallLeaves
			default:
				b.LeafSize = bambooNoLeaves
			}
		}
		tx.SetBlock(pos, b, nil)
	}
}
