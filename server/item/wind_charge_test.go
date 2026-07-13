package item

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestWindChargeLaunchVelocity(t *testing.T) {
	tests := []struct {
		name            string
		direction       mgl64.Vec3
		deviation       mgl64.Vec3
		shooterVelocity mgl64.Vec3
		onGround        bool
		want            mgl64.Vec3
	}{
		{
			name:            "stationary",
			direction:       mgl64.Vec3{0, 0, 1},
			shooterVelocity: mgl64.Vec3{},
			onGround:        true,
			want:            mgl64.Vec3{0, 0, 1.5},
		},
		{
			name:            "grounded shooter horizontal motion",
			direction:       mgl64.Vec3{0, 0, 1},
			shooterVelocity: mgl64.Vec3{0.2, 0.3, -0.1},
			onGround:        true,
			want:            mgl64.Vec3{0.2, 0, 1.4},
		},
		{
			name:            "airborne shooter full motion",
			direction:       mgl64.Vec3{0, 0, 1},
			shooterVelocity: mgl64.Vec3{0.2, 0.3, -0.1},
			onGround:        false,
			want:            mgl64.Vec3{0.2, 0.3, 1.4},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := windChargeLaunchVelocity(test.direction, test.deviation, test.shooterVelocity, test.onGround)
			if !got.ApproxEqualThreshold(test.want, 1e-12) {
				t.Fatalf("launch velocity = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWindChargeLaunchVelocityAddsDeviationAfterNormalisingDirection(t *testing.T) {
	direction := mgl64.Vec3{0, 0, 1}
	deviation := mgl64.Vec3{0.0172275, -0.0172275, 0}
	want := direction.Normalize().Add(deviation).Mul(1.5)

	got := windChargeLaunchVelocity(direction, deviation, mgl64.Vec3{}, true)
	if !got.ApproxEqualThreshold(want, 1e-12) {
		t.Fatalf("launch velocity = %v, want %v", got, want)
	}
}
