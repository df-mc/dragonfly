package session

import (
	"image/color"
	"math/rand"
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NetworkEncodeableEntity is an Entity where the save ID and network ID are not the same.
type NetworkEncodeableEntity interface {
	// NetworkEncodeEntity returns the network type ID of the entity. This is NOT the save ID.
	NetworkEncodeEntity() string
}

// entityHidden checks if a world.Entity is being explicitly hidden from the Session.
func (s *Session) entityHidden(e world.Entity) bool {
	s.entityMutex.RLock()
	_, ok := s.hiddenEntities[e]
	s.entityMutex.RUnlock()
	return ok
}

// ViewEntity ...
func (s *Session) ViewEntity(e world.Entity) {
	if s.entityRuntimeID(e) == selfEntityRuntimeID {
		s.ViewEntityState(e)
		return
	}
	if s.entityHidden(e) {
		return
	}
	var runtimeID uint64

	_, controllable := e.(Controllable)

	s.entityMutex.Lock()
	if id, ok := s.entityRuntimeIDs[e]; ok && controllable {
		runtimeID = id
	} else {
		s.currentEntityRuntimeID += 1
		runtimeID = s.currentEntityRuntimeID
		s.entityRuntimeIDs[e] = runtimeID
		s.entities[runtimeID] = e
	}
	s.entityMutex.Unlock()

	yaw, pitch := e.Rotation()
	metadata := s.parseEntityMetadata(e)

	id := e.EncodeEntity()
	switch v := e.(type) {
	case Controllable:
		actualPlayer := false

		sessionMu.Lock()
		for _, s := range sessions {
			if s.c.UUID() == v.UUID() {
				actualPlayer = true
				break
			}
		}
		sessionMu.Unlock()
		if !actualPlayer {
			s.writePacket(&packet.PlayerList{ActionType: packet.PlayerListActionAdd, Entries: []protocol.PlayerListEntry{{
				UUID:           v.UUID(),
				EntityUniqueID: int64(runtimeID),
				Username:       v.Name(),
				Skin:           skinToProtocol(v.Skin()),
			}}})
		}

		s.writePacket(&packet.AddPlayer{
			UUID:            v.UUID(),
			Username:        v.Name(),
			EntityUniqueID:  int64(runtimeID),
			EntityRuntimeID: runtimeID,
			Position:        vec64To32(e.Position()),
			EntityMetadata:  metadata,
			Pitch:           float32(pitch),
			Yaw:             float32(yaw),
			HeadYaw:         float32(yaw),
			Layers: []protocol.AbilityLayer{{
				Type:      protocol.AbilityLayerTypeBase,
				Abilities: protocol.AbilityCount - 1,
			}},
		})
		if !actualPlayer {
			s.writePacket(&packet.PlayerList{ActionType: packet.PlayerListActionRemove, Entries: []protocol.PlayerListEntry{{
				UUID: v.UUID(),
			}}})
		}
		return
	case *entity.Item:
		s.writePacket(&packet.AddItemActor{
			EntityUniqueID:  int64(runtimeID),
			EntityRuntimeID: runtimeID,
			Item:            instanceFromItem(v.Item()),
			Position:        vec64To32(v.Position()),
			Velocity:        vec64To32(v.Velocity()),
			EntityMetadata:  metadata,
		})
		return
	case *entity.FallingBlock:
		metadata[dataKeyVariant] = int32(world.BlockRuntimeID(v.Block()))
	case *entity.Text:
		metadata[dataKeyVariant] = int32(world.BlockRuntimeID(block.Air{}))
	}
	if v, ok := e.(NetworkEncodeableEntity); ok {
		id = v.NetworkEncodeEntity()
	}

	var vel mgl64.Vec3
	if v, ok := e.(interface{ Velocity() mgl64.Vec3 }); ok {
		vel = v.Velocity()
	}

	s.writePacket(&packet.AddActor{
		EntityUniqueID:  int64(runtimeID),
		EntityRuntimeID: runtimeID,
		EntityType:      id,
		EntityMetadata:  metadata,
		Position:        vec64To32(e.Position()),
		Velocity:        vec64To32(vel),
		Pitch:           float32(pitch),
		Yaw:             float32(yaw),
		HeadYaw:         float32(yaw),
	})
}

// HideEntity ...
func (s *Session) HideEntity(e world.Entity) {
	if s.entityRuntimeID(e) == selfEntityRuntimeID {
		return
	}

	s.entityMutex.Lock()
	id, ok := s.entityRuntimeIDs[e]
	if _, controllable := e.(Controllable); !controllable {
		delete(s.entityRuntimeIDs, e)
		delete(s.entities, id)
	}
	s.entityMutex.Unlock()
	if !ok {
		// The entity was already removed some other way. We don't need to send a packet.
		return
	}
	s.writePacket(&packet.RemoveActor{EntityUniqueID: int64(id)})
}

// ViewEntityMovement ...
func (s *Session) ViewEntityMovement(e world.Entity, pos mgl64.Vec3, yaw, pitch float64, onGround bool) {
	id := s.entityRuntimeID(e)
	if id == selfEntityRuntimeID || s.entityHidden(e) {
		return
	}

	flags := byte(0)
	if onGround {
		flags |= packet.MoveFlagOnGround
	}
	s.writePacket(&packet.MoveActorAbsolute{
		EntityRuntimeID: id,
		Position:        vec64To32(pos.Add(entityOffset(e))),
		Rotation:        vec64To32(mgl64.Vec3{pitch, yaw, yaw}),
		Flags:           flags,
	})
}

// ViewEntityVelocity ...
func (s *Session) ViewEntityVelocity(e world.Entity, velocity mgl64.Vec3) {
	if s.entityHidden(e) {
		return
	}
	s.writePacket(&packet.SetActorMotion{
		EntityRuntimeID: s.entityRuntimeID(e),
		Velocity:        vec64To32(velocity),
	})
}

// entityOffset returns the offset that entities have client-side.
func entityOffset(e world.Entity) mgl64.Vec3 {
	switch e.(type) {
	case Controllable:
		return mgl64.Vec3{0, 1.62}
	case *entity.Item:
		return mgl64.Vec3{0, 0.125}
	case *entity.FallingBlock, *entity.TNT:
		return mgl64.Vec3{0, 0.49, 0}
	}
	return mgl64.Vec3{}
}

// ViewTime ...
func (s *Session) ViewTime(time int) {
	s.writePacket(&packet.SetTime{Time: int32(time)})
}

// ViewEntityTeleport ...
func (s *Session) ViewEntityTeleport(e world.Entity, position mgl64.Vec3) {
	id := s.entityRuntimeID(e)
	if s.entityHidden(e) {
		return
	}

	yaw, pitch := e.Rotation()
	if id == selfEntityRuntimeID {
		s.chunkLoader.Move(position)
		s.teleportPos.Store(&position)
	}

	s.writePacket(&packet.SetActorMotion{EntityRuntimeID: id})
	switch e.(type) {
	case Controllable:
		s.writePacket(&packet.MovePlayer{
			EntityRuntimeID: id,
			Position:        vec64To32(position.Add(entityOffset(e))),
			Pitch:           float32(pitch),
			Yaw:             float32(yaw),
			HeadYaw:         float32(yaw),
			Mode:            packet.MoveModeTeleport,
		})
	default:
		s.writePacket(&packet.MoveActorAbsolute{
			EntityRuntimeID: id,
			Position:        vec64To32(position.Add(entityOffset(e))),
			Rotation:        vec64To32(mgl64.Vec3{pitch, yaw, yaw}),
			Flags:           packet.MoveFlagTeleport,
		})
	}
}

// ViewEntityItems ...
func (s *Session) ViewEntityItems(e world.Entity) {
	runtimeID := s.entityRuntimeID(e)
	if runtimeID == selfEntityRuntimeID || s.entityHidden(e) {
		// Don't view the items of the entity if the entity is the Controllable entity of the session.
		return
	}
	c, ok := e.(item.Carrier)
	if !ok {
		return
	}

	mainHand, offHand := c.HeldItems()

	// Show the main hand item.
	s.writePacket(&packet.MobEquipment{
		EntityRuntimeID: runtimeID,
		NewItem:         instanceFromItem(mainHand),
	})
	// Show the off-hand item.
	s.writePacket(&packet.MobEquipment{
		EntityRuntimeID: runtimeID,
		NewItem:         instanceFromItem(offHand),
		WindowID:        protocol.WindowIDOffHand,
	})
}

// ViewEntityArmour ...
func (s *Session) ViewEntityArmour(e world.Entity) {
	runtimeID := s.entityRuntimeID(e)
	if runtimeID == selfEntityRuntimeID || s.entityHidden(e) {
		// Don't view the items of the entity if the entity is the Controllable entity of the session.
		return
	}
	armoured, ok := e.(interface {
		Armour() *inventory.Armour
	})
	if !ok {
		return
	}

	inv := armoured.Armour()

	// Show the main hand item.
	s.writePacket(&packet.MobArmourEquipment{
		EntityRuntimeID: runtimeID,
		Helmet:          instanceFromItem(inv.Helmet()),
		Chestplate:      instanceFromItem(inv.Chestplate()),
		Leggings:        instanceFromItem(inv.Leggings()),
		Boots:           instanceFromItem(inv.Boots()),
	})
}

// ViewItemCooldown ...
func (s *Session) ViewItemCooldown(item world.Item, duration time.Duration) {
	name, _ := item.EncodeItem()
	s.writePacket(&packet.ClientStartItemCooldown{
		Category: strings.Split(name, ":")[1],
		Duration: int32(duration.Milliseconds() / 50),
	})
}

// ViewParticle ...
func (s *Session) ViewParticle(pos mgl64.Vec3, p world.Particle) {
	switch pa := p.(type) {
	case particle.DragonEggTeleport:
		xSign, ySign, zSign := 0, 0, 0
		if pa.Diff.X() < 0 {
			xSign = 1 << 24
		}
		if pa.Diff.Y() < 0 {
			ySign = 1 << 25
		}
		if pa.Diff.Z() < 0 {
			zSign = 1 << 26
		}

		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticlesDragonEgg,
			Position:  vec64To32(pos),
			EventData: int32((((((abs(pa.Diff.X()) << 16) | (abs(pa.Diff.Y()) << 8)) | abs(pa.Diff.Z())) | xSign) | ySign) | zSign),
		})
	case particle.Note:
		s.writePacket(&packet.BlockEvent{
			EventType: pa.Instrument.Int32(),
			EventData: int32(pa.Pitch),
			Position:  protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())},
		})
	case particle.HugeExplosion:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticlesExplosion,
			Position:  vec64To32(pos),
		})
	case particle.BoneMeal:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleCropGrowth,
			Position:  vec64To32(pos),
		})
	case particle.BlockForceField:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleDenyBlock,
			Position:  vec64To32(pos),
		})
	case particle.BlockBreak:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticlesDestroyBlock,
			Position:  vec64To32(pos),
			EventData: int32(world.BlockRuntimeID(pa.Block)),
		})
	case particle.PunchBlock:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticlesCrackBlock,
			Position:  vec64To32(pos),
			EventData: int32(world.BlockRuntimeID(pa.Block)) | (int32(pa.Face) << 24),
		})
	case particle.EndermanTeleportParticle:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticlesTeleport,
			Position:  vec64To32(pos),
		})
	case particle.Flame:
		if pa.Colour != (color.RGBA{}) {
			s.writePacket(&packet.LevelEvent{
				EventType: packet.LevelEventParticleLegacyEvent | 56,
				Position:  vec64To32(pos),
				EventData: nbtconv.Int32FromRGBA(pa.Colour),
			})
			return
		}
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 8,
			Position:  vec64To32(pos),
		})
	case particle.Evaporate:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticlesEvaporateWater,
			Position:  vec64To32(pos),
		})
	case particle.SnowballPoof:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 15,
			Position:  vec64To32(pos),
		})
	case particle.EggSmash:
		rid, meta, _ := world.ItemRuntimeID(item.Egg{})
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 14,
			EventData: (rid << 16) | int32(meta),
			Position:  vec64To32(pos),
		})
	case particle.Splash:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticlesPotionSplash,
			EventData: (int32(pa.Colour.A) << 24) | (int32(pa.Colour.R) << 16) | (int32(pa.Colour.G) << 8) | int32(pa.Colour.B),
			Position:  vec64To32(pos),
		})
	case particle.Effect:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 33,
			EventData: (int32(pa.Colour.A) << 24) | (int32(pa.Colour.R) << 16) | (int32(pa.Colour.G) << 8) | int32(pa.Colour.B),
			Position:  vec64To32(pos),
		})
	case particle.EntityFlame:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 18,
			Position:  vec64To32(pos),
		})
	case particle.Dust:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 32,
			Position:  vec64To32(pos),
			EventData: nbtconv.Int32FromRGBA(pa.Colour),
		})
	case particle.WaterDrip:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 27,
			Position:  vec64To32(pos),
		})
	case particle.LavaDrip:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 28,
			Position:  vec64To32(pos),
		})
	case particle.Lava:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventParticleLegacyEvent | 10,
			Position:  vec64To32(pos),
		})
	}
}

// tierToSoundEvent converts an item.ArmourTier to a sound event associated with equipping it.
func tierToSoundEvent(tier item.ArmourTier) uint32 {
	switch tier.(type) {
	case item.ArmourTierLeather:
		return packet.SoundEventEquipLeather
	case item.ArmourTierGold:
		return packet.SoundEventEquipGold
	case item.ArmourTierChain:
		return packet.SoundEventEquipChain
	case item.ArmourTierIron:
		return packet.SoundEventEquipIron
	case item.ArmourTierDiamond:
		return packet.SoundEventEquipDiamond
	case item.ArmourTierNetherite:
		return packet.SoundEventEquipNetherite
	}
	return packet.SoundEventEquipGeneric
}

// playSound plays a world.Sound at a position, disabling relative volume if set to true.
func (s *Session) playSound(pos mgl64.Vec3, t world.Sound, disableRelative bool) {
	pk := &packet.LevelSoundEvent{
		Position:              vec64To32(pos),
		EntityType:            ":",
		ExtraData:             -1,
		DisableRelativeVolume: disableRelative,
	}
	switch so := t.(type) {
	case sound.EquipItem:
		switch i := so.Item.(type) {
		case item.Helmet:
			pk.SoundType = tierToSoundEvent(i.Tier)
		case item.Chestplate:
			pk.SoundType = tierToSoundEvent(i.Tier)
		case item.Leggings:
			pk.SoundType = tierToSoundEvent(i.Tier)
		case item.Boots:
			pk.SoundType = tierToSoundEvent(i.Tier)
		case item.Elytra:
			pk.SoundType = packet.SoundEventEquipElytra
		default:
			pk.SoundType = packet.SoundEventEquipGeneric
		}
	case sound.Note:
		pk.SoundType = packet.SoundEventNote
		pk.ExtraData = (so.Instrument.Int32() << 8) | int32(so.Pitch)
	case sound.DoorCrash:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundZombieDoorCrash,
			Position:  vec64To32(pos),
		})
		return
	case sound.Explosion:
		pk.SoundType = packet.SoundEventExplode
	case sound.Thunder:
		pk.SoundType, pk.EntityType = packet.SoundEventThunder, "minecraft:lightning_bolt"
	case sound.Click:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundClick,
			Position:  vec64To32(pos),
		})
		return
	case sound.Pop:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundInfinityArrowPickup,
			Position:  vec64To32(pos),
		})
		return
	case sound.Teleport:
		pk.SoundType = packet.SoundEventTeleport
	case sound.ItemFrameAdd:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundAddItem,
			Position:  vec64To32(pos),
		})
		return
	case sound.ItemFrameRemove:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundItemFrameRemoveItem,
			Position:  vec64To32(pos),
		})
		return
	case sound.ItemFrameRotate:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundItemFrameRotateItem,
			Position:  vec64To32(pos),
		})
		return
	case sound.GhastWarning:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundGhastWarning,
			Position:  vec64To32(pos),
		})
		return
	case sound.GhastShoot:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundGhastFireball,
			Position:  vec64To32(pos),
		})
		return
	case sound.TNT:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundFuse,
			Position:  vec64To32(pos),
		})
		return
	case sound.FireworkLaunch:
		pk.SoundType = packet.SoundEventLaunch
	case sound.FireworkHugeBlast:
		pk.SoundType = packet.SoundEventLargeBlast
	case sound.FireworkBlast:
		pk.SoundType = packet.SoundEventBlast
	case sound.FireworkTwinkle:
		pk.SoundType = packet.SoundEventTwinkle
	case sound.FurnaceCrackle:
		pk.SoundType = packet.SoundEventFurnaceUse
	case sound.BlastFurnaceCrackle:
		pk.SoundType = packet.SoundEventBlastFurnaceUse
	case sound.SmokerCrackle:
		pk.SoundType = packet.SoundEventSmokerUse
	case sound.UseSpyglass:
		pk.SoundType = packet.SoundEventUseSpyglass
	case sound.StopUsingSpyglass:
		pk.SoundType = packet.SoundEventStopUsingSpyglass
	case sound.GoatHorn:
		switch so.Horn {
		case sound.Ponder():
			pk.SoundType = packet.SoundEventGoatCall0
		case sound.Sing():
			pk.SoundType = packet.SoundEventGoatCall1
		case sound.Seek():
			pk.SoundType = packet.SoundEventGoatCall2
		case sound.Feel():
			pk.SoundType = packet.SoundEventGoatCall3
		case sound.Admire():
			pk.SoundType = packet.SoundEventGoatCall4
		case sound.Call():
			pk.SoundType = packet.SoundEventGoatCall5
		case sound.Yearn():
			pk.SoundType = packet.SoundEventGoatCall6
		case sound.Dream():
			pk.SoundType = packet.SoundEventGoatCall7
		}
	case sound.FireExtinguish:
		pk.SoundType = packet.SoundEventExtinguishFire
	case sound.Ignite:
		pk.SoundType = packet.SoundEventIgnite
	case sound.Burning:
		pk.SoundType = packet.SoundEventPlayerHurtOnFire
	case sound.Drowning:
		pk.SoundType = packet.SoundEventPlayerHurtDrown
	case sound.Fall:
		pk.EntityType = "minecraft:player"
		if so.Distance > 4 {
			pk.SoundType = packet.SoundEventFallBig
			break
		}
		pk.SoundType = packet.SoundEventFallSmall
	case sound.Burp:
		pk.SoundType = packet.SoundEventBurp
	case sound.Door:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundOpenDoor,
			Position:  vec64To32(pos),
		})
		return
	case sound.Deny:
		pk.SoundType = packet.SoundEventDeny
	case sound.BlockPlace:
		pk.SoundType, pk.ExtraData = packet.SoundEventPlace, int32(world.BlockRuntimeID(so.Block))
	case sound.AnvilLand:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundAnvilLand,
			Position:  vec64To32(pos),
		})
		return
	case sound.AnvilUse:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundAnvilUsed,
			Position:  vec64To32(pos),
		})
		return
	case sound.AnvilBreak:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundAnvilBroken,
			Position:  vec64To32(pos),
		})
		return
	case sound.ChestClose:
		pk.SoundType = packet.SoundEventChestClosed
	case sound.ChestOpen:
		pk.SoundType = packet.SoundEventChestOpen
	case sound.BarrelClose:
		pk.SoundType = packet.SoundEventBarrelClose
	case sound.BarrelOpen:
		pk.SoundType = packet.SoundEventBarrelOpen
	case sound.BlockBreaking:
		pk.SoundType, pk.ExtraData = packet.SoundEventHit, int32(world.BlockRuntimeID(so.Block))
	case sound.ItemBreak:
		pk.SoundType = packet.SoundEventBreak
	case sound.ItemUseOn:
		pk.SoundType, pk.ExtraData = packet.SoundEventItemUseOn, int32(world.BlockRuntimeID(so.Block))
	case sound.Fizz:
		pk.SoundType = packet.SoundEventFizz
	case sound.GlassBreak:
		pk.SoundType = packet.SoundEventGlass
	case sound.Attack:
		pk.SoundType, pk.EntityType = packet.SoundEventAttackStrong, "minecraft:player"
		if !so.Damage {
			pk.SoundType = packet.SoundEventAttackNoDamage
		}
	case sound.BucketFill:
		if _, water := so.Liquid.(block.Water); water {
			pk.SoundType = packet.SoundEventBucketFillWater
			break
		}
		pk.SoundType = packet.SoundEventBucketFillLava
	case sound.BucketEmpty:
		if _, water := so.Liquid.(block.Water); water {
			pk.SoundType = packet.SoundEventBucketEmptyWater
			break
		}
		pk.SoundType = packet.SoundEventBucketEmptyLava
	case sound.BowShoot:
		pk.SoundType = packet.SoundEventBow
	case sound.ArrowHit:
		pk.SoundType = packet.SoundEventBowHit
	case sound.ItemThrow:
		pk.SoundType, pk.EntityType = packet.SoundEventThrow, "minecraft:player"
	case sound.LevelUp:
		pk.SoundType, pk.ExtraData = packet.SoundEventLevelUp, 0x10000000
	case sound.Experience:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundExperienceOrbPickup,
			Position:  vec64To32(pos),
		})
		return
	case sound.MusicDiscPlay:
		switch so.DiscType {
		case sound.Disc13():
			pk.SoundType = packet.SoundEventRecord13
		case sound.DiscCat():
			pk.SoundType = packet.SoundEventRecordCat
		case sound.DiscBlocks():
			pk.SoundType = packet.SoundEventRecordBlocks
		case sound.DiscChirp():
			pk.SoundType = packet.SoundEventRecordChirp
		case sound.DiscFar():
			pk.SoundType = packet.SoundEventRecordFar
		case sound.DiscMall():
			pk.SoundType = packet.SoundEventRecordMall
		case sound.DiscMellohi():
			pk.SoundType = packet.SoundEventRecordMellohi
		case sound.DiscStal():
			pk.SoundType = packet.SoundEventRecordStal
		case sound.DiscStrad():
			pk.SoundType = packet.SoundEventRecordStrad
		case sound.DiscWard():
			pk.SoundType = packet.SoundEventRecordWard
		case sound.Disc11():
			pk.SoundType = packet.SoundEventRecord11
		case sound.DiscWait():
			pk.SoundType = packet.SoundEventRecordWait
		case sound.DiscOtherside():
			pk.SoundType = packet.SoundEventRecordOtherside
		case sound.DiscPigstep():
			pk.SoundType = packet.SoundEventRecordPigstep
		case sound.Disc5():
			pk.SoundType = packet.SoundEventRecord5
		}
	case sound.MusicDiscEnd:
		pk.SoundType = packet.SoundEventRecordNull
	case sound.FireCharge:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventSoundBlazeFireball,
			Position:  vec64To32(pos),
		})
	case sound.ComposterEmpty:
		pk.SoundType = packet.SoundEventComposterEmpty
	case sound.ComposterFill:
		pk.SoundType = packet.SoundEventComposterFill
	case sound.ComposterFillLayer:
		pk.SoundType = packet.SoundEventComposterFillLayer
	case sound.ComposterReady:
		pk.SoundType = packet.SoundEventComposterReady
	}
	s.writePacket(pk)
}

// PlaySound plays a world.Sound to the client. The volume is not dependent on the distance to the source if it is a
// sound of the LevelSoundEvent packet.
func (s *Session) PlaySound(t world.Sound) {
	if s == Nop {
		return
	}
	s.playSound(entity.EyePosition(s.c), t, true)
}

// ViewSound ...
func (s *Session) ViewSound(pos mgl64.Vec3, soundType world.Sound) {
	s.playSound(pos, soundType, false)
}

// ViewFurnaceUpdate updates a furnace for the associated session based on previous times.
func (s *Session) ViewFurnaceUpdate(prevCookTime, cookTime, prevRemainingFuelTime, remainingFuelTime, prevMaxFuelTime, maxFuelTime time.Duration) {
	if prevCookTime != cookTime {
		s.writePacket(&packet.ContainerSetData{
			WindowID: byte(s.openedWindowID.Load()),
			Key:      packet.ContainerDataFurnaceTickCount,
			Value:    int32(cookTime.Milliseconds() / 50),
		})
	}

	if prevRemainingFuelTime != remainingFuelTime {
		s.writePacket(&packet.ContainerSetData{
			WindowID: byte(s.openedWindowID.Load()),
			Key:      packet.ContainerDataFurnaceLitTime,
			Value:    int32(remainingFuelTime.Milliseconds() / 50),
		})
	}

	if prevMaxFuelTime != maxFuelTime {
		s.writePacket(&packet.ContainerSetData{
			WindowID: byte(s.openedWindowID.Load()),
			Key:      packet.ContainerDataFurnaceLitDuration,
			Value:    int32(maxFuelTime.Milliseconds() / 50),
		})
	}
}

// ViewBlockUpdate ...
func (s *Session) ViewBlockUpdate(pos cube.Pos, b world.Block, layer int) {
	blockPos := protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}
	s.writePacket(&packet.UpdateBlock{
		Position:          blockPos,
		NewBlockRuntimeID: world.BlockRuntimeID(b),
		Flags:             packet.BlockUpdateNetwork,
		Layer:             uint32(layer),
	})
	if v, ok := b.(world.NBTer); ok {
		NBTData := v.EncodeNBT()
		NBTData["x"], NBTData["y"], NBTData["z"] = int32(pos.X()), int32(pos.Y()), int32(pos.Z())
		s.writePacket(&packet.BlockActorData{
			Position: blockPos,
			NBTData:  NBTData,
		})
	}
}

// ViewEntityAction ...
func (s *Session) ViewEntityAction(e world.Entity, a world.EntityAction) {
	switch act := a.(type) {
	case entity.SwingArmAction:
		if _, ok := e.(Controllable); ok {
			if s.entityRuntimeID(e) == selfEntityRuntimeID && s.swingingArm.Load() {
				return
			}
			s.writePacket(&packet.Animate{
				ActionType:      packet.AnimateActionSwingArm,
				EntityRuntimeID: s.entityRuntimeID(e),
			})
			return
		}
		s.writePacket(&packet.ActorEvent{
			EntityRuntimeID: s.entityRuntimeID(e),
			EventType:       packet.ActorEventStartAttacking,
		})
	case entity.HurtAction:
		s.writePacket(&packet.ActorEvent{
			EntityRuntimeID: s.entityRuntimeID(e),
			EventType:       packet.ActorEventHurt,
		})
	case entity.CriticalHitAction:
		s.writePacket(&packet.Animate{
			ActionType:      packet.AnimateActionCriticalHit,
			EntityRuntimeID: s.entityRuntimeID(e),
		})
	case entity.DeathAction:
		s.writePacket(&packet.ActorEvent{
			EntityRuntimeID: s.entityRuntimeID(e),
			EventType:       packet.ActorEventDeath,
		})
	case entity.PickedUpAction:
		s.writePacket(&packet.TakeItemActor{
			ItemEntityRuntimeID:  s.entityRuntimeID(e),
			TakerEntityRuntimeID: s.entityRuntimeID(act.Collector),
		})
	case entity.ArrowShakeAction:
		s.writePacket(&packet.ActorEvent{
			EntityRuntimeID: s.entityRuntimeID(e),
			EventType:       packet.ActorEventShake,
			EventData:       int32(act.Duration.Milliseconds() / 50),
		})
	case entity.FireworkExplosionAction:
		s.writePacket(&packet.ActorEvent{
			EntityRuntimeID: s.entityRuntimeID(e),
			EventType:       packet.ActorEventFireworksExplode,
		})
	case entity.EatAction:
		if user, ok := e.(item.User); ok {
			held, _ := user.HeldItems()
			it := held.Item()
			if held.Empty() {
				// This can happen sometimes if the user switches between items very quickly, so just ignore the action.
				return
			}
			if _, ok := it.(item.Consumable); !ok {
				// Not consumable, refer to the comment above.
				return
			}
			rid, meta, _ := world.ItemRuntimeID(it)
			s.writePacket(&packet.ActorEvent{
				EntityRuntimeID: s.entityRuntimeID(e),
				EventType:       packet.ActorEventFeed,
				// It's a little weird how the runtime ID is still shifted 16 bits to the left here, given the
				// runtime ID already includes the meta, but it seems to work.
				EventData: (rid << 16) | int32(meta),
			})
		}
	}
}

// ViewEntityState ...
func (s *Session) ViewEntityState(e world.Entity) {
	s.writePacket(&packet.SetActorData{
		EntityRuntimeID: s.entityRuntimeID(e),
		EntityMetadata:  s.parseEntityMetadata(e),
	})
}

// OpenBlockContainer ...
func (s *Session) OpenBlockContainer(pos cube.Pos) {
	if s.containerOpened.Load() && s.openedPos.Load() == pos {
		return
	}
	s.closeCurrentContainer()

	w := s.c.World()
	b := w.Block(pos)
	if container, ok := b.(block.Container); ok {
		s.openNormalContainer(container, pos)
		return
	}
	// We hit a special kind of window like beacons, which are not actually opened server-side.
	nextID := s.nextWindowID()
	s.containerOpened.Store(true)
	s.openedWindow.Store(inventory.New(1, nil))
	s.openedPos.Store(pos)

	var containerType byte
	switch b := b.(type) {
	case block.CraftingTable:
		containerType = 1
	case block.EnchantingTable:
		containerType = 3
	case block.Anvil:
		containerType = 5
	case block.Beacon:
		containerType = 13
	case block.Loom:
		containerType = 24
	case block.Grindstone:
		containerType = 26
	case block.Stonecutter:
		containerType = 29
	case block.SmithingTable:
		containerType = 33
	case block.EnderChest:
		b.AddViewer(w, pos)
		inv := s.c.EnderChestInventory()
		s.openedWindow.Store(inv)
		defer s.sendInv(inv, uint32(nextID))
	}
	s.openedContainerID.Store(uint32(containerType))
	s.writePacket(&packet.ContainerOpen{
		WindowID:                nextID,
		ContainerType:           containerType,
		ContainerPosition:       protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		ContainerEntityUniqueID: -1,
	})
}

const (
	containerTypeFurnace      = 2
	containerTypeBlastFurnace = 27
	containerTypeSmoker       = 28
)

// openNormalContainer opens a normal container that can hold items in it server-side.
func (s *Session) openNormalContainer(b block.Container, pos cube.Pos) {
	b.AddViewer(s, s.c.World(), pos)

	nextID := s.nextWindowID()
	s.containerOpened.Store(true)
	s.openedWindow.Store(b.Inventory())
	s.openedPos.Store(pos)

	var containerType byte
	switch b.(type) {
	case block.Furnace:
		containerType = containerTypeFurnace
	case block.BlastFurnace:
		containerType = containerTypeBlastFurnace
	case block.Smoker:
		containerType = containerTypeSmoker
	}

	s.writePacket(&packet.ContainerOpen{
		WindowID:                nextID,
		ContainerType:           containerType,
		ContainerPosition:       protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
		ContainerEntityUniqueID: -1,
	})
	s.sendInv(b.Inventory(), uint32(nextID))
}

// ViewSlotChange ...
func (s *Session) ViewSlotChange(slot int, newItem item.Stack) {
	if !s.containerOpened.Load() {
		return
	}
	if s.inTransaction.Load() {
		// Don't send slot changes to the player itself.
		return
	}
	s.writePacket(&packet.InventorySlot{
		WindowID: s.openedWindowID.Load(),
		Slot:     uint32(slot),
		NewItem:  instanceFromItem(newItem),
	})
}

// ViewBlockAction ...
func (s *Session) ViewBlockAction(pos cube.Pos, a world.BlockAction) {
	blockPos := protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}
	switch t := a.(type) {
	case block.OpenAction:
		s.writePacket(&packet.BlockEvent{
			Position:  blockPos,
			EventType: packet.BlockEventChangeChestState,
			EventData: 1,
		})
	case block.CloseAction:
		s.writePacket(&packet.BlockEvent{
			Position:  blockPos,
			EventType: packet.BlockEventChangeChestState,
		})
	case block.StartCrackAction:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventStartBlockCracking,
			Position:  vec64To32(pos.Vec3()),
			EventData: int32(65535 / (t.BreakTime.Seconds() * 20)),
		})
	case block.StopCrackAction:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventStopBlockCracking,
			Position:  vec64To32(pos.Vec3()),
			EventData: 0,
		})
	case block.ContinueCrackAction:
		s.writePacket(&packet.LevelEvent{
			EventType: packet.LevelEventUpdateBlockCracking,
			Position:  vec64To32(pos.Vec3()),
			EventData: int32(65535 / (t.BreakTime.Seconds() * 20)),
		})
	}
}

// ViewEmote ...
func (s *Session) ViewEmote(player world.Entity, emote uuid.UUID) {
	s.writePacket(&packet.Emote{
		EntityRuntimeID: s.entityRuntimeID(player),
		EmoteID:         emote.String(),
		Flags:           packet.EmoteFlagServerSide,
	})
}

// ViewSkin ...
func (s *Session) ViewSkin(e world.Entity) {
	switch v := e.(type) {
	case Controllable:
		s.writePacket(&packet.PlayerSkin{
			UUID: v.UUID(),
			Skin: skinToProtocol(v.Skin()),
		})
	}
}

// ViewWorldSpawn ...
func (s *Session) ViewWorldSpawn(pos cube.Pos) {
	blockPos := protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}
	s.writePacket(&packet.SetSpawnPosition{
		SpawnType:     packet.SpawnTypeWorld,
		Position:      blockPos,
		Dimension:     packet.DimensionOverworld,
		SpawnPosition: blockPos,
	})
}

// ViewWeather ...
func (s *Session) ViewWeather(raining, thunder bool) {
	pk := &packet.LevelEvent{
		EventType: packet.LevelEventStopRaining,
	}
	if raining {
		pk.EventType, pk.EventData = packet.LevelEventStartRaining, int32(rand.Intn(50000)+10000)
	}
	s.writePacket(pk)

	pk = &packet.LevelEvent{
		EventType: packet.LevelEventStopThunderstorm,
	}
	if thunder {
		pk.EventType, pk.EventData = packet.LevelEventStartThunderstorm, int32(rand.Intn(50000)+10000)
	}
	s.writePacket(pk)
}

// nextWindowID produces the next window ID for a new window. It is an int of 1-99.
func (s *Session) nextWindowID() byte {
	if s.openedWindowID.CAS(99, 1) {
		return 1
	}
	return byte(s.openedWindowID.Add(1))
}

// closeWindow closes the container window currently opened. If no window is open, closeWindow will do
// nothing.
func (s *Session) closeWindow() {
	if !s.containerOpened.CAS(true, false) {
		return
	}
	s.openedContainerID.Store(0)
	s.openedWindow.Store(inventory.New(1, nil))
	s.writePacket(&packet.ContainerClose{WindowID: byte(s.openedWindowID.Load())})
}

// entityRuntimeID returns the runtime ID of the entity passed.
//noinspection GoCommentLeadingSpace
func (s *Session) entityRuntimeID(e world.Entity) uint64 {
	s.entityMutex.RLock()
	//lint:ignore S1005 Double assignment is done explicitly to prevent panics.
	id, _ := s.entityRuntimeIDs[e]
	s.entityMutex.RUnlock()
	return id
}

// entityFromRuntimeID attempts to return an entity by its runtime ID. False is returned if no entity with the
// ID could be found.
func (s *Session) entityFromRuntimeID(id uint64) (world.Entity, bool) {
	s.entityMutex.RLock()
	e, ok := s.entities[id]
	s.entityMutex.RUnlock()
	return e, ok
}

// vec32To64 converts a mgl32.Vec3 to a mgl64.Vec3.
func vec32To64(vec3 mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{float64(vec3[0]), float64(vec3[1]), float64(vec3[2])}
}

// vec64To32 converts a mgl64.Vec3 to a mgl32.Vec3.
func vec64To32(vec3 mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(vec3[0]), float32(vec3[1]), float32(vec3[2])}
}

// boolByte returns 1 if the bool passed is true, or 0 if it is false.
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// abs ...
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
