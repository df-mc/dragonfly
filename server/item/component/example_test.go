package component_test

import (
	"fmt"
	"image"

	"github.com/df-mc/dragonfly/server/item/category"
	"github.com/df-mc/dragonfly/server/item/component"
	"github.com/df-mc/dragonfly/server/world"
)

var _ world.CustomItem = TestHelmet{}
var _ component.ComponentItem = TestHelmet{}

// TestHelmet is a custom item that demonstrates using the generated typed component
// structs via component.ComponentItem, with the SlotArmor named string type.
type TestHelmet struct{}

func (TestHelmet) EncodeItem() (name string, meta int16) {
	return "test:helmet", 0
}

func (TestHelmet) Name() string {
	return "Test Helmet"
}

func (TestHelmet) Texture() image.Image {
	return image.NewRGBA(image.Rect(0, 0, 16, 16))
}

func (TestHelmet) Category() category.Category {
	return category.Equipment()
}

func (TestHelmet) ItemComponents() []component.Component {
	return []component.Component{
		component.Wearable{
			Slot:                component.SlotArmorHead,
			Protection:          3,
			HidesPlayerLocation: false,
			Dispensable:         true,
		},
		component.Durability{
			MaxDurability: 200,
			DamageChance:  [2]int32{1, 100},
		},
		component.DamageAbsorption{
			AbsorbableCauses: []string{"fall", "explosion"},
		},
	}
}

func ExampleWearable() {
	h := TestHelmet{}
	for _, c := range h.ItemComponents() {
		if c.ComponentName() != "minecraft:wearable" {
			continue
		}
		data, _ := c.Encode()
		fmt.Printf("slot=%v\n", data["slot"])
		break
	}
	// Output:
	// slot=slot.armor.head
}
