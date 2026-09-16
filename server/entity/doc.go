// Package entity implements the entities that inhabit a world, along with the machinery to build new ones,
// both inside this package and in external modules.
//
// # Architecture
//
// Three types work together to form an entity:
//
//   - A world.EntityType is the identity of an entity kind. It creates the entity's live
//     form through its Open method and encodes/decodes its data to NBT.
//   - A world.EntityHandle is the persistent form of one entity. It owns the entity's world.EntityData and
//     outlives transactions, worlds and restarts.
//   - An Ent is the live form of an entity inside a single world.Tx. It is cheap and short-lived: it is
//     recreated through EntityType.Open for every transaction that touches the entity.
//
// Because an Ent only lives for one transaction, an entity's state must never be stored on it. All state is
// held by the entity's Behaviour, which is stored in world.EntityData.Data and travels with the handle. The
// Behaviour's Tick drives the entity and returns a *Movement describing how it moved, which the Ent sends to
// viewers.
//
// # Behaviours
//
// This package provides behaviours for common entity archetypes: PassiveBehaviour for entities that are
// pushed around by the environment, ProjectileBehaviour for entities that fly and hit, StationaryBehaviour
// for entities that stay put, and specialised behaviours built on them. Each behaviour is created from its
// XBehaviourConfig through New, and each config's Apply method stores a new behaviour in a
// world.EntityData, so configs may be passed to world.EntitySpawnOpts.New directly.
//
// Behaviours advertise optional abilities to the rest of the server by implementing extension interfaces,
// discovered by type assertion on the Behaviour:
//
//   - DamageableBehaviour lets an entity take damage without being alive. HurtEntity and Ent.Hurt
//     dispatch to it.
//   - ExplodableBehaviour lets an entity react to explosions. Ent.Explode dispatches to it.
//   - PortalTravelHandler and PortalTravelComputerProvider hook an entity into portal travel; embedding
//     BaseBehaviour provides the latter.
//   - LivingHurtHandler and LivingDeathHandler let a LivingBehaviour intercept damage and death.
//
// # Living entities
//
// A living entity is built by implementing LivingBehaviour: a Behaviour that exposes a
// LivingState holding its state: health, effects, armour, held items, air, fall distance and appearance.
// The entity type's Open method returns OpenLiving(tx, handle, data), and the resulting LivingEnt
// implements Living. Everything gated on Living then works out of the box: the full player attack path
// with enchantments and critical hits, projectile knock back, potion effects, attack immunity with the
// vanilla excess-damage rule, vanilla explosion damage, damage reduction, thorns and durability of worn
// armour, burning, drowning, suffocation, void and fall damage, the contact hooks of blocks stood in or
// on, and a death deferred to the next tick so drops can be spawned within a transaction. Worn armour and
// held items are shown to viewers, as are the scale, baby, variant and breathing state of the LivingState.
//
// AI is deliberately not part of this package: it is built in external
// modules on top of LivingBehaviour, which drive their entities by returning a Movement from Tick and by
// reporting motion through LivingEnt.UpdateFallState.
//
// # Building entities outside this package
//
// An external module defines its own world.EntityType and Behaviour. A behaviour that runs its own physics
// returns NewMovement(e, pos, vel, rot, onGround) from Tick to move its entity; to reuse this package's
// gravity, drag and collision it may run MovementComputer.TickMovement first and feed the resulting
// Position and Velocity into NewMovement. It broadcasts animations with Ent.PlayAction and resends metadata
// with Ent.UpdateState after changing state it reports. A behaviour contributes raw metadata values to
// viewers by implementing EncodeEntityMetadata(m map[uint32]any). Damage to
// arbitrary entities is dealt through HurtEntity, and DamageableEntity reports whether an entity can take
// damage at all. The world's Handler may change or cancel damage to any entity through HandleEntityHurt.
//
// The DecodeNBT method of a world.EntityType must always store a Behaviour in the world.EntityData passed,
// like its Open-time config does: an entity loaded from disk is otherwise opened without a Behaviour and
// panics on first use.
//
// For a server to spawn custom entity types, they must be part of its world.EntityRegistry:
//
//	reg := entity.DefaultRegistry.Config().New(append(entity.DefaultRegistry.Types(), custom.Type))
package entity
