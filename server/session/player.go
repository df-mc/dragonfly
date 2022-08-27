package session

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"net"
	"time"
	_ "unsafe" // Imported for compiler directives.
)

// StopShowingEntity stops showing a world.Entity to the Session. It will be completely invisible until a call to
// StartShowingEntity is made.
func (s *Session) StopShowingEntity(e world.Entity) {
	s.HideEntity(e)
	s.entityMutex.Lock()
	s.hiddenEntities[e] = struct{}{}
	s.entityMutex.Unlock()
}

// StartShowingEntity starts showing a world.Entity to the Session that was previously hidden using StopShowingEntity.
func (s *Session) StartShowingEntity(e world.Entity) {
	s.entityMutex.Lock()
	delete(s.hiddenEntities, e)
	s.entityMutex.Unlock()
	s.ViewEntity(e)
	s.ViewEntityState(e)
	s.ViewEntityItems(e)
	s.ViewEntityArmour(e)
}

// closeCurrentContainer closes the container the player might currently have open.
func (s *Session) closeCurrentContainer() {
	if !s.containerOpened.Load() {
		return
	}
	s.closeWindow()

	pos := s.openedPos.Load()
	w := s.c.World()
	b := w.Block(pos)
	if container, ok := b.(block.Container); ok {
		container.RemoveViewer(s, w, pos)
	} else if enderChest, ok := b.(block.EnderChest); ok {
		enderChest.RemoveViewer(w, pos)
	}
}

// SendRespawn spawns the Controllable entity of the session client-side in the world, provided it has died.
func (s *Session) SendRespawn(pos mgl64.Vec3) {
	s.writePacket(&packet.Respawn{
		Position:        vec64To32(pos.Add(entityOffset(s.c))),
		State:           packet.RespawnStateReadyToSpawn,
		EntityRuntimeID: selfEntityRuntimeID,
	})
}

// sendRecipes sends the current crafting recipes to the session.
func (s *Session) sendRecipes() {
	s.writePacket(&packet.CraftingData{Recipes: s.protocolRecipes(), ClearRecipes: true})
}

// sendInv sends the inventory passed to the client with the window ID.
func (s *Session) sendInv(inv *inventory.Inventory, windowID uint32) {
	pk := &packet.InventoryContent{
		WindowID: windowID,
		Content:  make([]protocol.ItemInstance, 0, s.inv.Size()),
	}
	for _, i := range inv.Slots() {
		pk.Content = append(pk.Content, instanceFromItem(i))
	}
	s.writePacket(pk)
}

// sendItem sends the item stack passed to the client with the window ID and slot passed.
func (s *Session) sendItem(item item.Stack, slot int, windowID uint32) {
	s.writePacket(&packet.InventorySlot{
		WindowID: windowID,
		Slot:     uint32(slot),
		NewItem:  instanceFromItem(item),
	})
}

const (
	craftingGridSizeSmall   = 4
	craftingGridSizeLarge   = 9
	craftingGridSmallOffset = 28
	craftingGridLargeOffset = 32
	craftingResult          = 50
)

const (
	containerAnvilInput            = 0
	containerAnvilMaterial         = 1
	containerSmithingInput         = 3
	containerSmithingMaterial      = 4
	containerArmour                = 6
	containerChest                 = 7
	containerBeacon                = 8
	containerFullInventory         = 12
	containerCraftingGrid          = 13
	containerEnchantingTableInput  = 21
	containerEnchantingTableLapis  = 22
	containerFurnaceFuel           = 23
	containerFurnaceResult         = 25
	containerFurnaceInput          = 24
	containerHotbar                = 27
	containerInventory             = 28
	containerOffHand               = 33
	containerLoomInput             = 40
	containerLoomDye               = 41
	containerLoomPattern           = 42
	containerBlastFurnaceInput     = 44
	containerSmokerInput           = 45
	containerGrindstoneFirstInput  = 49
	containerGrindstoneSecondInput = 50
	containerStonecutterInput      = 52
	containerBarrel                = 57
	containerCursor                = 58
	containerOutput                = 59
)

// smelter is an interface representing a block used to smelt items.
type smelter interface {
	// ResetExperience resets the collected experience of the smelter, and returns the amount of experience that was reset.
	ResetExperience() int
}

// invByID attempts to return an inventory by the ID passed. If found, the inventory is returned and the bool
// returned is true.
func (s *Session) invByID(id int32) (*inventory.Inventory, bool) {
	switch id {
	case containerCraftingGrid, containerOutput, containerCursor:
		// UI inventory.
		return s.ui, true
	case containerHotbar, containerInventory, containerFullInventory:
		// Hotbar 'inventory', rest of inventory, inventory when container is opened.
		return s.inv, true
	case containerOffHand:
		return s.offHand, true
	case containerArmour:
		// Armour inventory.
		return s.armour.Inventory(), true
	case containerChest:
		if s.containerOpened.Load() {
			b := s.c.World().Block(s.openedPos.Load())
			if _, chest := b.(block.Chest); chest {
				return s.openedWindow.Load(), true
			} else if _, enderChest := b.(block.EnderChest); enderChest {
				return s.openedWindow.Load(), true
			}
		}
	case containerBarrel:
		if s.containerOpened.Load() {
			if _, barrel := s.c.World().Block(s.openedPos.Load()).(block.Barrel); barrel {
				return s.openedWindow.Load(), true
			}
		}
	case containerBeacon:
		if s.containerOpened.Load() {
			if _, beacon := s.c.World().Block(s.openedPos.Load()).(block.Beacon); beacon {
				return s.ui, true
			}
		}
	case containerAnvilInput, containerAnvilMaterial:
		if s.containerOpened.Load() {
			if _, anvil := s.c.World().Block(s.openedPos.Load()).(block.Anvil); anvil {
				return s.ui, true
			}
		}
	case containerSmithingInput, containerSmithingMaterial:
		if s.containerOpened.Load() {
			if _, smithing := s.c.World().Block(s.openedPos.Load()).(block.SmithingTable); smithing {
				return s.ui, true
			}
		}
	case containerLoomInput, containerLoomDye, containerLoomPattern:
		if s.containerOpened.Load() {
			if _, loom := s.c.World().Block(s.openedPos.Load()).(block.Loom); loom {
				return s.ui, true
			}
		}
	case containerStonecutterInput:
		if s.containerOpened.Load() {
			if _, ok := s.c.World().Block(s.openedPos.Load()).(block.Stonecutter); ok {
				return s.ui, true
			}
		}
	case containerGrindstoneFirstInput, containerGrindstoneSecondInput:
		if s.containerOpened.Load() {
			if _, ok := s.c.World().Block(s.openedPos.Load()).(block.Grindstone); ok {
				return s.ui, true
			}
		}
	case containerEnchantingTableInput, containerEnchantingTableLapis:
		if s.containerOpened.Load() {
			if _, enchanting := s.c.World().Block(s.openedPos.Load()).(block.EnchantingTable); enchanting {
				return s.ui, true
			}
		}
	case containerFurnaceInput, containerFurnaceFuel, containerFurnaceResult, containerBlastFurnaceInput, containerSmokerInput:
		if s.containerOpened.Load() {
			if _, ok := s.c.World().Block(s.openedPos.Load()).(smelter); ok {
				return s.openedWindow.Load(), true
			}
		}
	}
	return nil, false
}

// Disconnect disconnects the client and ultimately closes the session. If the message passed is non-empty,
// it will be shown to the client.
func (s *Session) Disconnect(message string) {
	if s != Nop {
		s.writePacket(&packet.Disconnect{
			HideDisconnectionScreen: message == "",
			Message:                 message,
		})
		_ = s.conn.Flush()
	}
}

// SendSpeed sends the speed of the player in an UpdateAttributes packet, so that it is updated client-side.
func (s *Session) SendSpeed(speed float64) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			AttributeValue: protocol.AttributeValue{
				Name:  "minecraft:movement",
				Value: float32(speed),
				Max:   math.MaxFloat32,
			},
			Default: 0.1,
		}},
	})
}

// SendFood ...
func (s *Session) SendFood(food int, saturation, exhaustion float64) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.hunger",
					Value: float32(food),
					Max:   20,
				},
				Default: 20,
			},
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.saturation",
					Value: float32(saturation),
					Max:   20,
				},
				Default: 20,
			},
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.exhaustion",
					Value: float32(exhaustion),
					Max:   5,
				},
			},
		},
	})
}

// SendForm sends a form to the client of the connection. The Submit method of the form is called when the
// client submits the form.
func (s *Session) SendForm(f form.Form) {
	b, _ := json.Marshal(f)

	h := s.handlers[packet.IDModalFormResponse].(*ModalFormResponseHandler)
	id := h.currentID.Add(1)

	h.mu.Lock()
	if len(h.forms) > 10 {
		s.log.Debugf("SendForm %v: more than 10 active forms: dropping an existing one.", s.c.Name())
		for k := range h.forms {
			delete(h.forms, k)
			break
		}
	}
	h.forms[id] = f
	h.mu.Unlock()

	s.writePacket(&packet.ModalFormRequest{
		FormID:   id,
		FormData: b,
	})
}

// Transfer transfers the player to a server with the IP and port passed.
func (s *Session) Transfer(ip net.IP, port int) {
	s.writePacket(&packet.Transfer{
		Address: ip.String(),
		Port:    uint16(port),
	})
}

// SendGameMode sends the game mode of the Controllable entity of the session to the client. It makes sure the right
// flags are set to create the full game mode.
func (s *Session) SendGameMode(mode world.GameMode) {
	if s == Nop {
		return
	}

	id := int32(packet.GameTypeSurvival)
	if mode.AllowsFlying() && mode.CreativeInventory() {
		id = packet.GameTypeCreative
	}
	if !mode.Visible() && !mode.HasCollision() {
		id = packet.GameTypeSpectator
	}
	s.writePacket(&packet.SetPlayerGameType{GameType: id})
	s.sendAbilities()
}

// sendAbilities sends the abilities of the Controllable entity of the session to the client.
func (s *Session) sendAbilities() {
	mode, abilities := s.c.GameMode(), uint32(0)
	if mode.AllowsFlying() {
		abilities |= protocol.AbilityMayFly
		if s.c.Flying() {
			abilities |= protocol.AbilityFlying
		}
	}
	if !mode.HasCollision() {
		abilities |= protocol.AbilityNoClip
		defer s.c.StartFlying()
		// If the client is currently on the ground and turned to spectator mode, it will be unable to sprint during
		// flight. In order to allow this, we force the client to be flying through a MovePlayer packet.
		s.ViewEntityTeleport(s.c, s.c.Position())
	}
	if !mode.AllowsTakingDamage() {
		abilities |= protocol.AbilityInvulnerable
	}
	if mode.CreativeInventory() {
		abilities |= protocol.AbilityInstantBuild
	}
	if mode.AllowsEditing() {
		abilities |= protocol.AbilityBuild | protocol.AbilityMine
	}
	if mode.AllowsInteraction() {
		abilities |= protocol.AbilityDoorsAndSwitches | protocol.AbilityOpenContainers | protocol.AbilityAttackPlayers | protocol.AbilityAttackMobs
	}
	s.writePacket(&packet.UpdateAbilities{
		EntityUniqueID:     selfEntityRuntimeID,
		PlayerPermissions:  packet.PermissionLevelMember,
		CommandPermissions: packet.CommandPermissionLevelNormal,
		Layers: []protocol.AbilityLayer{ // TODO: Support customization of fly and walk speeds.
			{
				Type:      protocol.AbilityLayerTypeBase,
				Abilities: protocol.AbilityCount - 1,
				Values:    abilities,
				FlySpeed:  protocol.AbilityBaseFlySpeed,
				WalkSpeed: protocol.AbilityBaseWalkSpeed,
			},
		},
	})
}

// SendHealth sends the health and max health to the player.
func (s *Session) SendHealth(health *entity.HealthManager) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			AttributeValue: protocol.AttributeValue{
				Name:  "minecraft:health",
				Value: float32(math.Ceil(health.Health())),
				Max:   float32(math.Ceil(health.MaxHealth())),
			},
			Default: 20,
		}},
	})
}

// SendAbsorption sends the absorption value passed to the player.
func (s *Session) SendAbsorption(value float64) {
	maximum := value
	if math.Mod(value, 2) != 0 {
		maximum = value + 1
	}
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			AttributeValue: protocol.AttributeValue{
				Name:  "minecraft:absorption",
				Value: float32(math.Ceil(value)),
				Max:   float32(math.Ceil(maximum)),
			},
		}},
	})
}

// SendEffect sends an effects passed to the player.
func (s *Session) SendEffect(e effect.Effect) {
	s.SendEffectRemoval(e.Type())
	id, _ := effect.ID(e.Type())
	s.writePacket(&packet.MobEffect{
		EntityRuntimeID: selfEntityRuntimeID,
		Operation:       packet.MobEffectAdd,
		EffectType:      int32(id),
		Amplifier:       int32(e.Level() - 1),
		Particles:       !e.ParticlesHidden(),
		Duration:        int32(e.Duration() / (time.Second / 20)),
	})
}

// SendEffectRemoval sends the removal of an effect passed.
func (s *Session) SendEffectRemoval(e effect.Type) {
	id, ok := effect.ID(e)
	if !ok {
		panic(fmt.Sprintf("unregistered effect type %T", e))
	}
	s.writePacket(&packet.MobEffect{
		EntityRuntimeID: selfEntityRuntimeID,
		Operation:       packet.MobEffectRemove,
		EffectType:      int32(id),
	})
}

// SendGameRules sends all the provided game rules to the player. Once sent, they will be immediately updated
// on the client if they are valid.
func (s *Session) sendGameRules(gameRules []protocol.GameRule) {
	s.writePacket(&packet.GameRulesChanged{GameRules: gameRules})
}

// EnableCoordinates will either enable or disable coordinates for the player depending on the value given.
func (s *Session) EnableCoordinates(enable bool) {
	//noinspection SpellCheckingInspection
	s.sendGameRules([]protocol.GameRule{{Name: "showcoordinates", Value: enable}})
}

// EnableInstantRespawn will either enable or disable instant respawn for the player depending on the value given.
func (s *Session) EnableInstantRespawn(enable bool) {
	//noinspection SpellCheckingInspection
	s.sendGameRules([]protocol.GameRule{{Name: "doimmediaterespawn", Value: enable}})
}

// addToPlayerList adds the player of a session to the player list of this session. It will be shown in the
// in-game pause menu screen.
func (s *Session) addToPlayerList(session *Session) {
	c := session.c

	runtimeID := uint64(1)
	s.entityMutex.Lock()
	if session != s {
		s.currentEntityRuntimeID += 1
		runtimeID = s.currentEntityRuntimeID
	}
	s.entityRuntimeIDs[c] = runtimeID
	s.entities[runtimeID] = c
	s.entityMutex.Unlock()

	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionAdd,
		Entries: []protocol.PlayerListEntry{{
			UUID:           c.UUID(),
			EntityUniqueID: int64(runtimeID),
			Username:       c.Name(),
			XUID:           c.XUID(),
			Skin:           skinToProtocol(c.Skin()),
		}},
	})
}

// skinToProtocol converts a skin to its protocol representation.
func skinToProtocol(s skin.Skin) protocol.Skin {
	var animations []protocol.SkinAnimation
	for _, animation := range s.Animations {
		protocolAnim := protocol.SkinAnimation{
			ImageWidth:  uint32(animation.Bounds().Max.X),
			ImageHeight: uint32(animation.Bounds().Max.Y),
			ImageData:   animation.Pix,
			FrameCount:  float32(animation.FrameCount),
		}
		switch animation.Type() {
		case skin.AnimationHead:
			protocolAnim.AnimationType = protocol.SkinAnimationHead
		case skin.AnimationBody32x32:
			protocolAnim.AnimationType = protocol.SkinAnimationBody32x32
		case skin.AnimationBody128x128:
			protocolAnim.AnimationType = protocol.SkinAnimationBody128x128
		}
		protocolAnim.ExpressionType = uint32(animation.AnimationExpression)
		animations = append(animations, protocolAnim)
	}

	return protocol.Skin{
		PlayFabID:         s.PlayFabID,
		SkinID:            uuid.New().String(),
		SkinResourcePatch: s.ModelConfig.Encode(),
		SkinImageWidth:    uint32(s.Bounds().Max.X),
		SkinImageHeight:   uint32(s.Bounds().Max.Y),
		SkinData:          s.Pix,
		CapeImageWidth:    uint32(s.Cape.Bounds().Max.X),
		CapeImageHeight:   uint32(s.Cape.Bounds().Max.Y),
		CapeData:          s.Cape.Pix,
		SkinGeometry:      s.Model,
		PersonaSkin:       s.Persona,
		CapeID:            uuid.New().String(),
		FullID:            uuid.New().String(),
		Animations:        animations,
		Trusted:           true,
	}
}

// removeFromPlayerList removes the player of a session from the player list of this session. It will no
// longer be shown in the in-game pause menu screen.
func (s *Session) removeFromPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	delete(s.entities, s.entityRuntimeIDs[c])
	delete(s.entityRuntimeIDs, c)
	s.entityMutex.Unlock()

	s.writePacket(&packet.PlayerList{
		ActionType: packet.PlayerListActionRemove,
		Entries: []protocol.PlayerListEntry{{
			UUID: c.UUID(),
		}},
	})
}

// HandleInventories starts handling the inventories of the Controllable entity of the session. It sends packets when
// slots in the inventory are changed.
func (s *Session) HandleInventories() (inv, offHand, enderChest *inventory.Inventory, armour *inventory.Armour, heldSlot *atomic.Uint32) {
	s.inv = inventory.New(36, func(slot int, item item.Stack) {
		if s.c == nil {
			return
		}
		if slot == int(s.heldSlot.Load()) {
			for _, viewer := range s.c.World().Viewers(s.c.Position()) {
				viewer.ViewEntityItems(s.c)
			}
		}
		if !s.inTransaction.Load() {
			s.sendItem(item, slot, protocol.WindowIDInventory)
		}
	})
	s.offHand = inventory.New(1, func(slot int, item item.Stack) {
		if s.c == nil {
			return
		}
		for _, viewer := range s.c.World().Viewers(s.c.Position()) {
			viewer.ViewEntityItems(s.c)
		}
		if !s.inTransaction.Load() {
			i, _ := s.offHand.Item(0)
			s.writePacket(&packet.InventoryContent{
				WindowID: protocol.WindowIDOffHand,
				Content: []protocol.ItemInstance{
					instanceFromItem(i),
				},
			})
		}
	})
	s.enderChest = inventory.New(27, func(slot int, item item.Stack) {
		if s.c == nil {
			return
		}
		if !s.inTransaction.Load() {
			if _, ok := s.c.World().Block(s.openedPos.Load()).(block.EnderChest); ok {
				s.ViewSlotChange(slot, item)
			}
		}
	})
	s.armour = inventory.NewArmour(func(slot int, item item.Stack) {
		if s.c == nil {
			return
		}
		for _, viewer := range s.c.World().Viewers(s.c.Position()) {
			viewer.ViewEntityArmour(s.c)
		}
		if !s.inTransaction.Load() {
			s.sendItem(item, slot, protocol.WindowIDArmour)
		}
	})
	return s.inv, s.offHand, s.enderChest, s.armour, s.heldSlot
}

// SetHeldSlot sets the currently held hotbar slot.
func (s *Session) SetHeldSlot(slot int) error {
	if slot > 8 {
		return fmt.Errorf("slot exceeds hotbar range 0-8: slot is %v", slot)
	}

	s.heldSlot.Store(uint32(slot))

	for _, viewer := range s.c.World().Viewers(s.c.Position()) {
		viewer.ViewEntityItems(s.c)
	}

	mainHand, _ := s.c.HeldItems()
	s.writePacket(&packet.MobEquipment{
		EntityRuntimeID: selfEntityRuntimeID,
		NewItem:         instanceFromItem(mainHand),
		InventorySlot:   byte(slot),
		HotBarSlot:      byte(slot),
	})
	return nil
}

// UpdateHeldSlot updates the held slot of the Session to the slot passed. It also verifies that the item in that slot
// matches an expected item stack.
func (s *Session) UpdateHeldSlot(slot int, expected item.Stack) error {
	// The slot that the player might have selected must be within the hotbar: The held item cannot be in a
	// different place in the inventory.
	if slot > 8 {
		return fmt.Errorf("new held slot exceeds hotbar range 0-8: slot is %v", slot)
	}
	if s.heldSlot.Load() == uint32(slot) {
		// Old slot was the same as new slot, so don't do anything.
		return nil
	}
	// The user swapped changed held slots so stop using item right away.
	s.c.ReleaseItem()

	s.heldSlot.Store(uint32(slot))

	clientSideItem := expected
	actual, _ := s.inv.Item(slot)

	// The item the client claims to have must be identical to the one we have registered server-side.
	if !clientSideItem.Equal(actual) {
		// Only ever debug these as they are frequent and expected to happen whenever client and server get
		// out of sync.
		s.log.Debugf("failed processing packet from %v (%v): failed changing held slot: client-side item must be identical to server-side item, but got differences: client: %v vs server: %v", s.conn.RemoteAddr(), s.c.Name(), clientSideItem, actual)
	}
	for _, viewer := range s.c.World().Viewers(s.c.Position()) {
		viewer.ViewEntityItems(s.c)
	}
	return nil
}

// SendExperience sends the experience level and progress from the given experience manager to the player.
func (s *Session) SendExperience(e *entity.ExperienceManager) {
	level, progress := e.Level(), e.Progress()
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.level",
					Value: float32(level),
					Max:   float32(math.MaxInt32),
				},
			},
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.experience",
					Value: float32(progress),
					Max:   1,
				},
			},
		},
	})
}

// protocolRecipes returns all recipes as protocol recipes.
func (s *Session) protocolRecipes() []protocol.Recipe {
	recipes := make([]protocol.Recipe, 0, len(recipe.Recipes()))
	for index, i := range recipe.Recipes() {
		networkID := uint32(index) + 1
		s.recipes[networkID] = i

		switch i := i.(type) {
		case recipe.Shapeless:
			recipes = append(recipes, &protocol.ShapelessRecipe{
				RecipeID:        uuid.New().String(),
				Priority:        int32(i.Priority()),
				Input:           stacksToIngredientItems(i.Input()),
				Output:          stacksToRecipeStacks(i.Output()),
				Block:           i.Block(),
				RecipeNetworkID: networkID,
			})
		case recipe.Shaped:
			recipes = append(recipes, &protocol.ShapedRecipe{
				RecipeID:        uuid.New().String(),
				Priority:        int32(i.Priority()),
				Width:           int32(i.Shape().Width()),
				Height:          int32(i.Shape().Height()),
				Input:           stacksToIngredientItems(i.Input()),
				Output:          stacksToRecipeStacks(i.Output()),
				Block:           i.Block(),
				RecipeNetworkID: networkID,
			})
		}
	}
	return recipes
}

// stackFromItem converts an item.Stack to its network ItemStack representation.
func stackFromItem(it item.Stack) protocol.ItemStack {
	if it.Empty() {
		return protocol.ItemStack{}
	}

	var blockRuntimeID uint32
	if b, ok := it.Item().(world.Block); ok {
		blockRuntimeID = world.BlockRuntimeID(b)
	}

	rid, meta, _ := world.ItemRuntimeID(it.Item())

	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     rid,
			MetadataValue: uint32(meta),
		},
		HasNetworkID:   true,
		Count:          uint16(it.Count()),
		BlockRuntimeID: int32(blockRuntimeID),
		NBTData:        nbtconv.WriteItem(it, false),
	}
}

// stackToItem converts a network ItemStack representation back to an item.Stack.
func stackToItem(it protocol.ItemStack) item.Stack {
	t, ok := world.ItemByRuntimeID(it.NetworkID, int16(it.MetadataValue))
	if !ok {
		t = block.Air{}
	}
	if it.BlockRuntimeID > 0 {
		// It shouldn't matter if it (for whatever reason) wasn't able to get the block runtime ID,
		// since on the next line, we assert that the block is an item. If it didn't succeed, it'll
		// return air anyway.
		b, _ := world.BlockByRuntimeID(uint32(it.BlockRuntimeID))
		if t, ok = b.(world.Item); !ok {
			t = block.Air{}
		}
	}
	//noinspection SpellCheckingInspection
	if nbter, ok := t.(world.NBTer); ok && len(it.NBTData) != 0 {
		t = nbter.DecodeNBT(it.NBTData).(world.Item)
	}
	s := item.NewStack(t, int(it.Count))
	return nbtconv.ReadItem(it.NBTData, &s)
}

// instanceFromItem converts an item.Stack to its network ItemInstance representation.
func instanceFromItem(it item.Stack) protocol.ItemInstance {
	return protocol.ItemInstance{
		StackNetworkID: item_id(it),
		Stack:          stackFromItem(it),
	}
}

// stacksToRecipeStacks converts a list of item.Stacks to their protocol representation with damage stripped for recipes.
func stacksToRecipeStacks(inputs []item.Stack) []protocol.ItemStack {
	items := make([]protocol.ItemStack, 0, len(inputs))
	for _, i := range inputs {
		items = append(items, deleteDamage(stackFromItem(i)))
	}
	return items
}

// stacksToIngredientItems converts a list of item.Stacks to recipe ingredient items used over the network.
func stacksToIngredientItems(inputs []item.Stack) []protocol.RecipeIngredientItem {
	items := make([]protocol.RecipeIngredientItem, 0, len(inputs))
	for _, i := range inputs {
		if i.Empty() {
			items = append(items, protocol.RecipeIngredientItem{})
			continue
		}
		rid, meta, ok := world.ItemRuntimeID(i.Item())
		if !ok {
			panic("should never happen")
		}
		if _, ok = i.Value("variants"); ok {
			meta = math.MaxInt16 // Used to indicate that the item has multiple selectable variants.
		}
		items = append(items, protocol.RecipeIngredientItem{
			NetworkID:     rid,
			MetadataValue: int32(meta),
			Count:         int32(i.Count()),
		})
	}
	return items
}

// creativeItems returns all creative inventory items as protocol item stacks.
func creativeItems() []protocol.CreativeItem {
	it := make([]protocol.CreativeItem, 0, len(creative.Items()))
	for index, i := range creative.Items() {
		it = append(it, protocol.CreativeItem{
			CreativeItemNetworkID: uint32(index) + 1,
			Item:                  deleteDamage(stackFromItem(i)),
		})
	}
	return it
}

// deleteDamage strips the damage from a protocol item.
func deleteDamage(st protocol.ItemStack) protocol.ItemStack {
	delete(st.NBTData, "Damage")
	return st
}

// protocolToSkin converts protocol.Skin to skin.Skin.
func protocolToSkin(sk protocol.Skin) (s skin.Skin, err error) {
	if sk.SkinID == "" {
		return skin.Skin{}, fmt.Errorf("SkinID must not be an empty string")
	}

	s = skin.New(int(sk.SkinImageWidth), int(sk.SkinImageHeight))
	s.Persona = sk.PersonaSkin
	s.Pix = sk.SkinData
	s.Model = sk.SkinGeometry
	s.PlayFabID = sk.PlayFabID

	s.Cape = skin.NewCape(int(sk.CapeImageWidth), int(sk.CapeImageHeight))
	s.Cape.Pix = sk.CapeData

	m := make(map[string]any)
	if err = json.Unmarshal(sk.SkinGeometry, &m); err != nil {
		return skin.Skin{}, fmt.Errorf("SkinGeometry was not a valid JSON string: %v", err)
	}

	if s.ModelConfig, err = skin.DecodeModelConfig(sk.SkinResourcePatch); err != nil {
		return skin.Skin{}, fmt.Errorf("SkinResourcePatch was not a valid JSON string: %v", err)
	}

	for _, anim := range sk.Animations {
		var t skin.AnimationType
		switch anim.AnimationType {
		case protocol.SkinAnimationHead:
			t = skin.AnimationHead
		case protocol.SkinAnimationBody32x32:
			t = skin.AnimationBody32x32
		case protocol.SkinAnimationBody128x128:
			t = skin.AnimationBody128x128
		default:
			return skin.Skin{}, fmt.Errorf("invalid animation type: %v", anim.AnimationType)
		}

		animation := skin.NewAnimation(int(anim.ImageWidth), int(anim.ImageHeight), int(anim.ExpressionType), t)
		animation.FrameCount = int(anim.FrameCount)
		animation.Pix = anim.ImageData

		s.Animations = append(s.Animations, animation)
	}
	return
}

// The following functions use the go:linkname directive in order to make sure the item.byID and item.toID
// functions do not need to be exported.

//go:linkname item_id github.com/df-mc/dragonfly/server/item.id
//noinspection ALL
func item_id(s item.Stack) int32

//go:linkname world_add github.com/df-mc/dragonfly/server/world.add
//noinspection ALL
func world_add(e world.Entity, w *world.World)
