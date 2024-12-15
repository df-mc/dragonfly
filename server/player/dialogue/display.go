package dialogue

import (
	"encoding/json"
	"github.com/go-gl/mathgl/mgl64"
)

// DisplaySettings holds optional fields that change the way the dialogue,
// particularly the entity shown in it, is displayed.
type DisplaySettings struct {
	// EntityScale specifies the scale of the entity displayed in the dialogue.
	EntityScale mgl64.Vec3
	// EntityOffset specifies the offset of the entity shown in the dialogue.
	EntityOffset mgl64.Vec3
	// EntityRotation is the rotation of the entity shown in the dialogue. This
	// rotation functions a bit differently to the normal entity rotation in
	// Minecraft: The values are still degrees, but pitch (rot[1]) values are
	// whole-body pitch instead of head-specific, and rot[2] is whole-body roll.
	EntityRotation mgl64.Vec3
}

// MarshalJSON encodes the DisplaySettings to JSON.
func (d DisplaySettings) MarshalJSON() ([]byte, error) {
	// Yaw and pitch are swapped in this display.
	d.EntityRotation[0], d.EntityRotation[1] = d.EntityRotation[1], d.EntityRotation[0]-32
	m := map[string]any{
		// Translate needs to be multiplied by -32 to get a rough block
		// equivalent.
		"translate": d.EntityOffset.Mul(-32),
		// Entity is rotated by 32 degrees by default.
		"rotate": d.EntityRotation,
		"scale":  [3]float64{1, 1, 1},
	}
	if (d.EntityScale != mgl64.Vec3{}) {
		m["scale"] = d.EntityScale
	}
	return json.Marshal(m)
}
