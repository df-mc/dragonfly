package player

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

// TestDropOversizedStack verifies that every item of a stack larger than its maximum count is dropped. An item entity
// holds at most a maximum sized stack and discards the rest, so dropping such a stack as one entity destroyed the
// remainder while reporting that all of it had been dropped.
func TestDropOversizedStack(t *testing.T) {
	tests := []struct {
		name       string
		count      int
		wantItems  int
		wantStacks int
	}{
		{name: "stack within the maximum", count: 30, wantItems: 30, wantStacks: 1},
		{name: "stack at the maximum", count: 64, wantItems: 64, wantStacks: 1},
		{name: "stack above the maximum", count: 128, wantItems: 128, wantStacks: 2},
		{name: "stack far above the maximum", count: 200, wantItems: 200, wantStacks: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
			defer w.Close()

			pos := mgl64.Vec3{0, 64, 0}
			conf := Config{Name: "test", UUID: uuid.New(), Position: pos, GameMode: world.GameModeSurvival}
			handle := world.EntitySpawnOpts{Position: pos}.New(Type, conf)

			var reported, items, stacks int
			w.Do(func(tx *world.Tx) {
				p := tx.AddEntity(handle).(*Player)

				reported = p.Drop(item.NewStack(item.Diamond{}, tt.count))

				for e := range tx.Entities() {
					b, ok := e.(interface{ Behaviour() entity.Behaviour })
					if !ok {
						continue
					}
					if it, ok := b.Behaviour().(*entity.ItemBehaviour); ok {
						items, stacks = items+it.Item().Count(), stacks+1
					}
				}
			}).Wait(context.Background())

			if items != tt.wantItems {
				t.Errorf("dropped %v items, want %v: %v were destroyed", items, tt.wantItems, tt.wantItems-items)
			}
			if stacks != tt.wantStacks {
				t.Errorf("dropped %v item entities, want %v", stacks, tt.wantStacks)
			}
			if reported != tt.wantItems {
				t.Errorf("Drop() = %v, want %v", reported, tt.wantItems)
			}
		})
	}
}
