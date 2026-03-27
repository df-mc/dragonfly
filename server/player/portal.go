package player

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	worldportal "github.com/df-mc/dragonfly/server/world/portal"
	"github.com/go-gl/mathgl/mgl64"
)

const netherPortalUseTicks = 80

// EnterNetherPortal marks the player as being inside a Nether portal this tick.
func (p *Player) EnterNetherPortal(pos cube.Pos, axis cube.Axis) {
	if time.Now().Before(p.portalCooldownUntil) {
		return
	}
	p.inNetherPortal = true
	p.netherPortalPos = pos
	p.netherPortalAxis = axis
}

// EnterEndPortal marks the player as being inside an End portal this tick.
func (p *Player) EnterEndPortal(pos cube.Pos) {
	if time.Now().Before(p.portalCooldownUntil) {
		return
	}
	p.inEndPortal = true
	p.endPortalPos = pos
}

// EnterEndGateway marks the player as being inside an End gateway this tick.
func (p *Player) EnterEndGateway(pos cube.Pos) {
	if time.Now().Before(p.portalCooldownUntil) {
		return
	}
	p.inEndGateway = true
	p.endGatewayPos = pos
}

func (p *Player) processPortals() bool {
	if p.inEndGateway {
		return p.travelThroughEndGateway()
	}
	if p.inEndPortal {
		return p.travelThroughEndPortal()
	}
	if !p.inNetherPortal {
		p.netherPortalTicks = 0
		return false
	}

	p.netherPortalTicks++
	if _, ok := p.Effect(effect.Nausea); !ok {
		p.AddEffect(effect.New(effect.Nausea, 1, time.Second*5).WithoutParticles())
	}
	if p.netherPortalTicks < netherPortalUseTicks {
		return false
	}
	return p.travelThroughNetherPortal()
}

func (p *Player) travelThroughEndGateway() bool {
	if p.tx.World().Dimension() != world.End {
		p.inEndGateway = false
		return false
	}
	target := p.tx.World().Spawn().Vec3Middle()
	p.teleport(target)
	p.finishPortalTransfer()
	return true
}

func (p *Player) travelThroughNetherPortal() bool {
	dest := p.tx.World().PortalDestination(world.Nether)
	if dest == nil || dest == p.tx.World() {
		p.netherPortalTicks = 0
		return false
	}

	sourcePos := cube.PosFromVec3(p.Position())
	targetPos := scaledNetherTarget(sourcePos, p.tx.World().Dimension(), dest.Range())
	handle := p.tx.RemoveEntity(p)
	if handle == nil {
		p.netherPortalTicks = 0
		return false
	}

	transferred := false
	<-dest.Exec(func(tx *world.Tx) {
		portalInfo, ok := worldportal.FindOrCreateNetherPortal(tx, targetPos, 128)
		destPos := targetPos.Vec3()
		if ok {
			destPos = netherPortalExitPosition(portalInfo, targetPos)
		}
		np := tx.AddEntity(handle).(*Player)
		np.teleport(destPos)
		np.finishPortalTransfer()
		transferred = true
	})
	return transferred
}

func (p *Player) travelThroughEndPortal() bool {
	dest := p.tx.World().PortalDestination(world.End)
	if dest == nil || dest == p.tx.World() {
		return false
	}

	target := dest.Spawn()
	if p.tx.World().Dimension() == world.End {
		target = dest.PlayerSpawn(p.UUID())
	}

	handle := p.tx.RemoveEntity(p)
	if handle == nil {
		return false
	}

	transferred := false
	<-dest.Exec(func(tx *world.Tx) {
		if tx.World().Dimension() == world.End {
			worldportal.EnsureEndEntryFeatures(tx)
		}
		np := tx.AddEntity(handle).(*Player)
		np.teleport(target.Vec3())
		np.finishPortalTransfer()
		transferred = true
	})
	return transferred
}

func (p *Player) finishPortalTransfer() {
	p.inNetherPortal = false
	p.inEndPortal = false
	p.inEndGateway = false
	p.netherPortalTicks = 0
	p.portalCooldownUntil = time.Now().Add(4 * time.Second)
	p.RemoveEffect(effect.Nausea)
}

func scaledNetherTarget(pos cube.Pos, source world.Dimension, destRange cube.Range) cube.Pos {
	x, z := pos.X(), pos.Z()
	if source == world.Nether {
		x *= 8
		z *= 8
	} else {
		x = floorDiv(x, 8)
		z = floorDiv(z, 8)
	}
	y := clamp(pos.Y(), destRange.Min()+1, destRange.Max()-1)
	return cube.Pos{x, y, z}
}

func netherPortalExitPosition(n worldportal.Nether, fallback cube.Pos) mgl64.Vec3 {
	positions := n.Positions()
	if len(positions) == 0 {
		return fallback.Vec3()
	}

	minX, maxX := positions[0].X(), positions[0].X()
	minY := positions[0].Y()
	minZ, maxZ := positions[0].Z(), positions[0].Z()
	for _, pos := range positions[1:] {
		minX = min(minX, pos.X())
		maxX = max(maxX, pos.X())
		minY = min(minY, pos.Y())
		minZ = min(minZ, pos.Z())
		maxZ = max(maxZ, pos.Z())
	}

	x := float64(minX+maxX+1) / 2
	z := float64(minZ+maxZ+1) / 2
	if n.Axis() == cube.X {
		x = float64(minX) + 0.5
	} else {
		z = float64(minZ) + 0.5
	}
	return mgl64.Vec3{x, float64(minY), z}
}

func floorDiv(x, y int) int {
	if y == 0 {
		panic("division by zero")
	}
	q := x / y
	r := x % y
	if r != 0 && ((r < 0) != (y < 0)) {
		q--
	}
	return q
}

func clamp(v, low, high int) int {
	return int(math.Max(float64(low), math.Min(float64(high), float64(v))))
}
