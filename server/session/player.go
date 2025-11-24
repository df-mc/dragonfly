package session

import (
	"encoding/json"
	"fmt"
	"image/color"
	"maps"
	"math"
	"net"
	"slices"
	"time"
	_ "unsafe" // Imported for compiler directives.

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/player/debug"
	"github.com/df-mc/dragonfly/server/player/dialogue"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/player/hud"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// StopShowingEntity stops showing a world.Entity to the Session. It will be completely invisible until a call to
// StartShowingEntity is made.
func (s *Session) StopShowingEntity(e world.Entity) {
	s.entityMutex.Lock()
	_, ok := s.hiddenEntities[e.H().UUID()]
	if !ok {
		s.hiddenEntities[e.H().UUID()] = struct{}{}
	}
	s.entityMutex.Unlock()

	if !ok {
		s.HideEntity(e)
	}
}

// StartShowingEntity starts showing a world.Entity to the Session that was previously hidden using StopShowingEntity.
func (s *Session) StartShowingEntity(e world.Entity) {
	s.entityMutex.Lock()
	_, ok := s.hiddenEntities[e.H().UUID()]
	if ok {
		delete(s.hiddenEntities, e.H().UUID())
	}
	s.entityMutex.Unlock()

	if ok {
		s.ViewEntity(e)
		s.ViewEntityState(e)
		s.ViewEntityItems(e)
		s.ViewEntityArmour(e)
	}
}

// closeCurrentContainer closes the container the player might currently have open.
func (s *Session) closeCurrentContainer(tx *world.Tx) {
	if !s.containerOpened.Load() {
		return
	}
	s.closeWindow()

	pos := *s.openedPos.Load()
	b := tx.Block(pos)
	if container, ok := b.(block.Container); ok {
		container.RemoveViewer(s, tx, pos)
	} else if enderChest, ok := b.(block.EnderChest); ok {
		enderChest.RemoveViewer(tx, pos)
	}
}

// SendRespawn spawns the Controllable entity of the session client-side in the world, provided it has died.
func (s *Session) SendRespawn(pos mgl64.Vec3, c Controllable) {
	s.writePacket(&packet.Respawn{
		Position:        vec64To32(pos.Add(entityOffset(c))),
		State:           packet.RespawnStateReadyToSpawn,
		EntityRuntimeID: selfEntityRuntimeID,
	})
}

// sendBiomes sends all the vanilla biomes to the session.
func (s *Session) sendBiomes() {
	definitions, stringList := world.BiomeDefinitions()
	s.writePacket(&packet.BiomeDefinitionList{
		BiomeDefinitions: definitions,
		StringList:       stringList,
	})
}

// sendRecipes sends the current crafting recipes to the session.
func (s *Session) sendRecipes() {
	recipes := make([]protocol.Recipe, 0, len(recipe.Recipes()))
	potionRecipes := make([]protocol.PotionRecipe, 0)
	potionContainerChange := make([]protocol.PotionContainerChangeRecipe, 0)

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
		case recipe.SmithingTransform:
			input, output := stacksToIngredientItems(i.Input()), stacksToRecipeStacks(i.Output())
			recipes = append(recipes, &protocol.SmithingTransformRecipe{
				RecipeID:        uuid.New().String(),
				Base:            input[0],
				Addition:        input[1],
				Template:        input[2],
				Result:          output[0],
				Block:           i.Block(),
				RecipeNetworkID: networkID,
			})
		case recipe.SmithingTrim:
			input := stacksToIngredientItems(i.Input())
			recipes = append(recipes, &protocol.SmithingTrimRecipe{
				RecipeID:        uuid.New().String(),
				Base:            input[0],
				Addition:        input[1],
				Template:        input[2],
				Block:           i.Block(),
				RecipeNetworkID: networkID,
			})
		case recipe.Furnace:
			recipes = append(recipes, &protocol.FurnaceRecipe{
				InputType: stackFromItem(i.Input()[0].(item.Stack)).ItemType,
				Output:    stackFromItem(i.Output()[0]),
				Block:     i.Block(),
			})
		case recipe.Potion:
			inputRuntimeID, inputMeta, _ := world.ItemRuntimeID(i.Input()[0].(item.Stack).Item())
			reagentRuntimeID, reagentMeta, _ := world.ItemRuntimeID(i.Input()[1].(item.Stack).Item())
			outputRuntimeID, outputMeta, _ := world.ItemRuntimeID(i.Output()[0].Item())

			potionRecipes = append(potionRecipes, protocol.PotionRecipe{
				InputPotionID:        inputRuntimeID,
				InputPotionMetadata:  int32(inputMeta),
				ReagentItemID:        reagentRuntimeID,
				ReagentItemMetadata:  int32(reagentMeta),
				OutputPotionID:       outputRuntimeID,
				OutputPotionMetadata: int32(outputMeta),
			})

		case recipe.PotionContainerChange:
			inputRuntimeID, _, _ := world.ItemRuntimeID(i.Input()[0].(item.Stack).Item())
			reagentRuntimeID, _, _ := world.ItemRuntimeID(i.Input()[1].(item.Stack).Item())
			outputRuntimeID, _, _ := world.ItemRuntimeID(i.Output()[0].Item())

			potionContainerChange = append(potionContainerChange, protocol.PotionContainerChangeRecipe{
				InputItemID:   inputRuntimeID,
				ReagentItemID: reagentRuntimeID,
				OutputItemID:  outputRuntimeID,
			})
		}
	}
	s.writePacket(&packet.CraftingData{Recipes: recipes, PotionRecipes: potionRecipes, PotionContainerChangeRecipes: potionContainerChange, ClearRecipes: true})
}

// sendArmourTrimData sends the armour trim data.
func (s *Session) sendArmourTrimData() {
	var trimPatterns []protocol.TrimPattern
	var trimMaterials []protocol.TrimMaterial

	for _, t := range item.SmithingTemplates() {
		if t == item.TemplateNetheriteUpgrade() {
			continue
		}
		name, _ := item.SmithingTemplate{Template: t}.EncodeItem()
		trimPatterns = append(trimPatterns, protocol.TrimPattern{
			ItemName:  name,
			PatternID: t.String(),
		})
	}

	for _, i := range item.ArmourTrimMaterials() {
		if material, ok := i.(item.ArmourTrimMaterial); ok {
			name, _ := i.EncodeItem()

			trimMaterials = append(trimMaterials, protocol.TrimMaterial{
				MaterialID: material.TrimMaterial(),
				Colour:     material.MaterialColour(),
				ItemName:   name,
			})
		}
	}

	s.writePacket(&packet.TrimData{Patterns: trimPatterns, Materials: trimMaterials})
}

// sendInv sends the inventory passed to the client with the window ID.
func (s *Session) sendInv(inv *inventory.Inventory, windowID uint32) {
	pk := &packet.InventoryContent{
		WindowID: windowID,
		Content:  make([]protocol.ItemInstance, 0, inv.Size()),
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

// smelter is an interface representing a block used to smelt items.
type smelter interface {
	// ResetExperience resets the collected experience of the smelter, and returns the amount of experience that was reset.
	ResetExperience() int
}

// invByID attempts to return an inventory by the ID passed. If found, the inventory is returned and the bool
// returned is true.
func (s *Session) invByID(id int32, tx *world.Tx) (*inventory.Inventory, bool) {
	switch id {
	case protocol.ContainerCraftingInput, protocol.ContainerCreatedOutput, protocol.ContainerCursor:
		// UI inventory.
		return s.ui, true
	case protocol.ContainerHotBar, protocol.ContainerInventory, protocol.ContainerCombinedHotBarAndInventory:
		// Hotbar 'inventory', rest of inventory, inventory when container is opened.
		return s.inv, true
	case protocol.ContainerOffhand:
		return s.offHand, true
	case protocol.ContainerArmor:
		// Armour inventory.
		return s.armour.Inventory(), true
	default:
		if !s.containerOpened.Load() {
			return nil, false
		}
		switch id {
		case protocol.ContainerLevelEntity:
			return s.openedWindow.Load(), true
		case protocol.ContainerBarrel:
			if _, barrel := tx.Block(*s.openedPos.Load()).(block.Barrel); barrel {
				return s.openedWindow.Load(), true
			}
		case protocol.ContainerBeaconPayment:
			if _, beacon := tx.Block(*s.openedPos.Load()).(block.Beacon); beacon {
				return s.ui, true
			}
		case protocol.ContainerBrewingStandInput, protocol.ContainerBrewingStandResult, protocol.ContainerBrewingStandFuel:
			if _, brewingStand := tx.Block(*s.openedPos.Load()).(block.BrewingStand); brewingStand {
				return s.openedWindow.Load(), true
			}
		case protocol.ContainerAnvilInput, protocol.ContainerAnvilMaterial:
			if _, anvil := tx.Block(*s.openedPos.Load()).(block.Anvil); anvil {
				return s.ui, true
			}
		case protocol.ContainerSmithingTableTemplate, protocol.ContainerSmithingTableInput, protocol.ContainerSmithingTableMaterial:
			if _, smithing := tx.Block(*s.openedPos.Load()).(block.SmithingTable); smithing {
				return s.ui, true
			}
		case protocol.ContainerLoomInput, protocol.ContainerLoomDye, protocol.ContainerLoomMaterial:
			if _, loom := tx.Block(*s.openedPos.Load()).(block.Loom); loom {
				return s.ui, true
			}
		case protocol.ContainerStonecutterInput:
			if _, ok := tx.Block(*s.openedPos.Load()).(block.Stonecutter); ok {
				return s.ui, true
			}
		case protocol.ContainerGrindstoneInput, protocol.ContainerGrindstoneAdditional:
			if _, ok := tx.Block(*s.openedPos.Load()).(block.Grindstone); ok {
				return s.ui, true
			}
		case protocol.ContainerEnchantingInput, protocol.ContainerEnchantingMaterial:
			if _, enchanting := tx.Block(*s.openedPos.Load()).(block.EnchantingTable); enchanting {
				return s.ui, true
			}
		case protocol.ContainerFurnaceIngredient, protocol.ContainerFurnaceFuel, protocol.ContainerFurnaceResult,
			protocol.ContainerBlastFurnaceIngredient, protocol.ContainerSmokerIngredient:
			if _, ok := tx.Block(*s.openedPos.Load()).(smelter); ok {
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
		_ = s.conn.WritePacket(&packet.Disconnect{
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
			DefaultMax: math.MaxFloat32,
			Default:    0.1,
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
				DefaultMax: 20,
				Default:    20,
			},
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.saturation",
					Value: float32(saturation),
					Max:   20,
				},
				DefaultMax: 20,
				Default:    20,
			},
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.exhaustion",
					Value: float32(exhaustion),
					Max:   5,
				},
				DefaultMax: 5,
			},
		},
	})
}

// SendDialogue sends an NPC dialogue to the client of the connection. The Submit method of the dialogue is
// called when the client interacts with a button in the dialogue.
func (s *Session) SendDialogue(d dialogue.Dialogue, e world.Entity) {
	b, _ := json.Marshal(d)

	h := s.handlers[packet.IDNPCRequest].(*NPCRequestHandler)
	h.dialogue = d
	h.entityRuntimeID = s.entityRuntimeID(e)

	metadata := s.parseEntityMetadata(e)
	metadata[protocol.EntityDataKeyHasNPC] = uint8(1)

	disp := d.Display()
	disp.EntityOffset = disp.EntityOffset.Add(entityOffset(e))
	display, _ := json.Marshal(map[string]any{"portrait_offsets": disp})
	metadata[protocol.EntityDataKeyNPCData] = string(display)

	s.writePacket(&packet.SetActorData{
		EntityRuntimeID: h.entityRuntimeID,
		EntityMetadata:  metadata,
	})
	s.writePacket(&packet.NPCDialogue{
		EntityUniqueID: h.entityRuntimeID,
		ActionType:     packet.NPCDialogueActionOpen,
		Dialogue:       d.Body(),
		SceneName:      "default",
		NPCName:        d.Title(),
		ActionJSON:     string(b),
	})
}

func (s *Session) CloseDialogue() {
	h := s.handlers[packet.IDNPCRequest].(*NPCRequestHandler)
	if h.entityRuntimeID == 0 {
		return
	}

	s.writePacket(&packet.NPCDialogue{
		EntityUniqueID: h.entityRuntimeID,
		ActionType:     packet.NPCDialogueActionClose,
	})
	h.entityRuntimeID = 0
}

// SendForm sends a form to the client of the connection. The Submit method of the form is called when the
// client submits the form.
func (s *Session) SendForm(f form.Form) {
	b, _ := json.Marshal(f)

	h := s.handlers[packet.IDModalFormResponse].(*ModalFormResponseHandler)
	id := h.currentID.Add(1)

	h.mu.Lock()
	if len(h.forms) > 10 {
		s.conf.Log.Debug("SendForm: more than 10 active forms: dropping an existing one")
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

// CloseForm closes any forms that the player currently has open. If the player has no forms open, nothing
// happens.
func (s *Session) CloseForm() {
	s.writePacket(&packet.ClientBoundCloseForm{})
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
func (s *Session) SendGameMode(c Controllable) {
	if s == Nop {
		return
	}
	s.writePacket(&packet.SetPlayerGameType{GameType: gameTypeFromMode(c.GameMode())})
	s.SendAbilities(c)
}

// SendAbilities sends the abilities of the Controllable entity of the session to the client.
func (s *Session) SendAbilities(c Controllable) {
	mode, abilities := c.GameMode(), uint32(0)
	if mode.AllowsFlying() {
		abilities |= protocol.AbilityMayFly
		if c.Flying() {
			abilities |= protocol.AbilityFlying
		}
	}
	if !mode.HasCollision() {
		abilities |= protocol.AbilityNoClip
		defer c.StartFlying()
		// If the client is currently on the ground and turned to spectator mode, it will be unable to sprint during
		// flight. In order to allow this, we force the client to be flying through a MovePlayer packet.
		s.ViewEntityTeleport(c, c.Position())
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
	s.writePacket(&packet.UpdateAbilities{AbilityData: protocol.AbilityData{
		EntityUniqueID:     selfEntityRuntimeID,
		PlayerPermissions:  packet.PermissionLevelMember,
		CommandPermissions: packet.CommandPermissionLevelNormal,
		Layers: []protocol.AbilityLayer{
			{
				Type:             protocol.AbilityLayerTypeBase,
				Abilities:        protocol.AbilityCount - 1,
				Values:           abilities,
				FlySpeed:         float32(c.FlightSpeed()),
				VerticalFlySpeed: float32(c.VerticalFlightSpeed()),
				WalkSpeed:        protocol.AbilityBaseWalkSpeed,
			},
		},
	}})
}

// SendHealth sends the health and max health to the player.
func (s *Session) SendHealth(health, max, absorption float64) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			AttributeValue: protocol.AttributeValue{
				Name:  "minecraft:health",
				Value: float32(math.Ceil(health)),
				Max:   float32(math.Ceil(max)),
			},
			DefaultMax: 20,
			Default:    20,
		}, {
			AttributeValue: protocol.AttributeValue{
				Name:  "minecraft:absorption",
				Value: float32(math.Ceil(absorption)),
				Max:   float32(math.MaxFloat32),
			},
			DefaultMax: float32(math.MaxFloat32),
		}},
	})
}

// SendEffect sends an effects passed to the player.
func (s *Session) SendEffect(e effect.Effect) {
	s.SendEffectRemoval(e.Type())
	id, _ := effect.ID(e.Type())
	dur := e.Duration() / (time.Second / 20)
	if e.Infinite() {
		dur = -1
	}
	s.writePacket(&packet.MobEffect{
		EntityRuntimeID: selfEntityRuntimeID,
		Operation:       packet.MobEffectAdd,
		EffectType:      int32(id),
		Amplifier:       int32(e.Level() - 1),
		Particles:       !e.ParticlesHidden(),
		Duration:        int32(dur),
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

// HandleInventories starts handling the inventories of the Controllable entity of the session. It sends packets when
// slots in the inventory are changed.
func (s *Session) HandleInventories(tx *world.Tx, c Controllable, inv, offHand, enderChest, ui *inventory.Inventory, armour *inventory.Armour, heldSlot *uint32) {
	s.inv = inv
	s.inv.SlotFunc(s.broadcastInvFunc(tx, c))
	s.offHand = offHand
	s.offHand.SlotFunc(s.broadcastOffHandFunc(tx, c))
	s.enderChest = enderChest
	s.enderChest.SlotFunc(s.broadcastEnderChestFunc(tx, c))
	s.armour = armour
	s.armour.Inventory().SlotFunc(s.broadcastArmourFunc(tx, c))
	s.ui = ui
	s.ui.SlotFunc(s.uiInventoryFunc(tx, c))
	s.heldSlot = heldSlot
}

func (s *Session) broadcastInvFunc(tx *world.Tx, c Controllable) inventory.SlotFunc {
	return func(slot int, _, after item.Stack) {
		if slot == int(*s.heldSlot) {
			for _, viewer := range tx.Viewers(c.Position()) {
				viewer.ViewEntityItems(c)
			}
		}
		if !s.inTransaction.Load() {
			s.sendItem(after, slot, protocol.WindowIDInventory)
		}
	}
}

func (s *Session) broadcastEnderChestFunc(tx *world.Tx, _ Controllable) inventory.SlotFunc {
	return func(slot int, _, after item.Stack) {
		if !s.inTransaction.Load() {
			if _, ok := tx.Block(*s.openedPos.Load()).(block.EnderChest); ok {
				s.ViewSlotChange(slot, after)
			}
		}
	}
}

func (s *Session) broadcastOffHandFunc(tx *world.Tx, c Controllable) inventory.SlotFunc {
	return func(slot int, _, after item.Stack) {
		for _, viewer := range tx.Viewers(c.Position()) {
			viewer.ViewEntityItems(c)
		}
		if !s.inTransaction.Load() {
			i, _ := s.offHand.Item(0)
			s.writePacket(&packet.InventoryContent{
				WindowID: protocol.WindowIDOffHand,
				Content:  []protocol.ItemInstance{instanceFromItem(i)},
			})
		}
	}
}

func (s *Session) broadcastArmourFunc(tx *world.Tx, c Controllable) inventory.SlotFunc {
	return func(slot int, before, after item.Stack) {
		if !s.inTransaction.Load() {
			s.sendItem(after, slot, protocol.WindowIDArmour)
		}
		if before.Comparable(after) && before.Empty() == after.Empty() {
			// Only send armour if the item type actually changed.
			return
		}
		for _, viewer := range tx.Viewers(c.Position()) {
			viewer.ViewEntityArmour(c)
		}
	}
}

// uiInventoryFunc handles an update to the UI inventory, used for updating enchantment options and possibly more
// in the future.
func (s *Session) uiInventoryFunc(tx *world.Tx, c Controllable) inventory.SlotFunc {
	return func(slot int, _, after item.Stack) {
		if slot == enchantingInputSlot && s.containerOpened.Load() {
			pos := *s.openedPos.Load()
			if _, enchanting := tx.Block(pos).(block.EnchantingTable); enchanting {
				s.sendEnchantmentOptions(tx, c, pos, after)
			}
		}
		s.sendInv(s.ui, protocol.WindowIDUI)
	}
}

// SendHeldSlot sends the currently held hotbar slot.
func (s *Session) SendHeldSlot(slot int, c Controllable, force bool) {
	if s.changingSlot.Load() && !force {
		return
	}
	mainHand, _ := c.HeldItems()
	s.writePacket(&packet.MobEquipment{
		EntityRuntimeID: selfEntityRuntimeID,
		NewItem:         instanceFromItem(mainHand),
		InventorySlot:   byte(slot),
		HotBarSlot:      byte(slot),
	})
}

// VerifyAndSetHeldSlot verifies if the slot passed is a valid hotbar slot and
// if the expected item.Stack is in it. Afterwards, it changes the held slot
// of the player.
func (s *Session) VerifyAndSetHeldSlot(slot int, expected item.Stack, c Controllable) error {
	if err := s.VerifySlot(slot, expected); err != nil {
		return err
	}
	s.changingSlot.Store(true)
	defer s.changingSlot.Store(false)
	return c.SetHeldSlot(slot)
}

// VerifySlot verifies if the slot passed is a valid hotbar slot and if the
// expected item.Stack is in it.
func (s *Session) VerifySlot(slot int, expected item.Stack) error {
	// The slot that the player might have selected must be within the hotbar:
	// The held item cannot be in a different place in the inventory.
	if slot < 0 || slot > 8 {
		return fmt.Errorf("slot exceeds hotbar range 0-8: slot is %v", slot)
	}
	clientSideItem := expected
	actual, _ := s.inv.Item(slot)

	// The item the client claims to have must be identical to the one we have
	// registered server-side.
	if !clientSideItem.Equal(actual) {
		s.sendItem(actual, slot, protocol.WindowIDInventory)
		// Only ever debug these as they are frequent and expected to happen
		// whenever client and server get out of sync.
		s.conf.Log.Debug("verify slot: client-side item was not equal to server-side item", "client-held", clientSideItem.String(), "server-held", actual.String())
	}
	return nil
}

// SendExperience sends the experience level and progress from the given experience manager to the player.
func (s *Session) SendExperience(level int, progress float64) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.level",
					Value: float32(level),
					Max:   float32(math.MaxInt32),
				},
				DefaultMax: float32(math.MaxInt32),
			},
			{
				AttributeValue: protocol.AttributeValue{
					Name:  "minecraft:player.experience",
					Value: float32(progress),
					Max:   1,
				},
				DefaultMax: 1,
			},
		},
	})
}

// SendChargeItemComplete sends a packet to indicate that the item charging process has been completed.
func (s *Session) SendChargeItemComplete() {
	s.writePacket(&packet.ActorEvent{
		EntityRuntimeID: selfEntityRuntimeID,
		EventType:       packet.ActorEventFinishedChargingItem,
	})
}

// ShowHudElement shows a HUD element to the player if it is not already shown. If the element is waiting to
// be hidden, it will be removed from the updates and remain visible to the player.
func (s *Session) ShowHudElement(e hud.Element) {
	s.hudMu.Lock()
	defer s.hudMu.Unlock()

	if _, ok := s.hiddenHud[e]; ok {
		s.hudUpdates[e] = true
	} else if _, ok = s.hudUpdates[e]; ok {
		delete(s.hudUpdates, e)
	}
}

// HideHudElement hides a HUD element from the player if it is not already hidden. If the element is waiting
// to be shown, it will be removed from the updates and remain hidden from the player.
func (s *Session) HideHudElement(e hud.Element) {
	s.hudMu.Lock()
	defer s.hudMu.Unlock()

	if _, ok := s.hiddenHud[e]; !ok {
		s.hudUpdates[e] = false
	} else if _, ok = s.hudUpdates[e]; ok {
		delete(s.hudUpdates, e)
	}
}

// HudElementHidden checks if a HUD element is currently hidden from the player.
func (s *Session) HudElementHidden(e hud.Element) bool {
	s.hudMu.RLock()
	defer s.hudMu.RUnlock()

	if _, ok := s.hiddenHud[e]; ok {
		return true
	}
	vis, ok := s.hudUpdates[e]
	return ok && !vis
}

// SendHudUpdates sends any pending HUD updates to the player. The updates are batched to reduce the number
// of packets being sent. Up to 2 packets will be sent, one for showing elements and one for hiding elements.
func (s *Session) SendHudUpdates() {
	s.hudMu.Lock()
	defer s.hudMu.Unlock()
	if len(s.hudUpdates) == 0 {
		return
	}
	var show, hide []int32
	for e, vis := range s.hudUpdates {
		if vis {
			show = append(show, int32(e.Uint8()))
			delete(s.hiddenHud, e)
		} else {
			hide = append(hide, int32(e.Uint8()))
			s.hiddenHud[e] = struct{}{}
		}
	}
	s.hudUpdates = make(map[hud.Element]bool)
	if len(show) > 0 {
		s.writePacket(&packet.SetHud{Elements: show, Visibility: packet.HudVisibilityReset})
	}
	if len(hide) > 0 {
		s.writePacket(&packet.SetHud{Elements: hide, Visibility: packet.HudVisibilityHide})
	}
}

// AddDebugShape adds a debug shape to be rendered to the player. If the shape already exists, it will be
// updated with the new information.
func (s *Session) AddDebugShape(shape debug.Shape) {
	s.debugShapesAdd <- shape
}

// RemoveDebugShape removes a debug shape from the player by its unique identifier.
func (s *Session) RemoveDebugShape(shape debug.Shape) {
	s.debugShapesMu.RLock()
	defer s.debugShapesMu.RUnlock()

	if _, ok := s.debugShapes[shape.ShapeID()]; ok {
		s.debugShapesRemove <- shape.ShapeID()
	}
}

// VisibleDebugShapes returns a slice of all debug shapes that are currently being shown to the player.
func (s *Session) VisibleDebugShapes() []debug.Shape {
	s.debugShapesMu.RLock()
	defer s.debugShapesMu.RUnlock()

	return slices.Collect(maps.Values(s.debugShapes))
}

// RemoveAllDebugShapes removes all rendered debug shapes from the player, as well as any shapes that have
// not yet been rendered.
func (s *Session) RemoveAllDebugShapes() {
	s.debugShapesMu.Lock()
	defer s.debugShapesMu.Unlock()

	s.debugShapesAdd = make(chan debug.Shape, 256)
	for id := range s.debugShapes {
		s.debugShapesRemove <- id
	}
}

// SendDebugShapes sends any pending additions/removals of debug shapes to the player. Shapes should be sent
// every tick to allow for batching and time-efficient updates.
func (s *Session) SendDebugShapes() {
	s.debugShapesMu.Lock()
	defer s.debugShapesMu.Unlock()

	if len(s.debugShapesAdd) == 0 && len(s.debugShapesRemove) == 0 {
		return
	}

	shapes := make([]packet.DebugDrawerShape, 0, len(s.debugShapesAdd)+len(s.debugShapesRemove))
loop:
	for {
		select {
		case shape := <-s.debugShapesAdd:
			s.debugShapes[shape.ShapeID()] = shape
			shapes = append(shapes, s.debugShapeToProtocol(shape))
		case id := <-s.debugShapesRemove:
			delete(s.debugShapes, id)
			shapes = append(shapes, packet.DebugDrawerShape{NetworkID: uint64(id)})
		default:
			break loop
		}
	}
	s.writePacket(&packet.DebugDrawer{Shapes: shapes})
}

// debugShapeToProtocol converts a debug shape to its protocol representation. It also provides defaults
// for some fields such as colour, scale and other per-shape properties.
func (s *Session) debugShapeToProtocol(shape debug.Shape) packet.DebugDrawerShape {
	ps := packet.DebugDrawerShape{NetworkID: uint64(shape.ShapeID())}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	switch shape := shape.(type) {
	case *debug.Arrow:
		ps.Type = protocol.Option(uint8(packet.ScriptDebugShapeArrow))
		ps.Colour = protocol.Option(valueOrDefault(shape.Colour, white))
		ps.Location = protocol.Option(vec64To32(shape.Position))
		ps.LineEndLocation = protocol.Option(vec64To32(shape.EndPosition))
		ps.ArrowHeadLength = protocol.Option(valueOrDefault(float32(shape.HeadLength), 1))
		ps.ArrowHeadRadius = protocol.Option(valueOrDefault(float32(shape.HeadRadius), 0.5))
		ps.Segments = protocol.Option(valueOrDefault(uint8(shape.HeadSegments), 4))
	case *debug.Box:
		ps.Type = protocol.Option(uint8(packet.ScriptDebugShapeBox))
		ps.Colour = protocol.Option(valueOrDefault(shape.Colour, white))
		ps.BoxBound = protocol.Option(valueOrDefault(vec64To32(shape.Bounds), mgl32.Vec3{1, 1, 1}))
		ps.Location = protocol.Option(vec64To32(shape.Position))
		ps.Scale = protocol.Option(valueOrDefault(float32(shape.Scale), 1))
	case *debug.Circle:
		ps.Type = protocol.Option(uint8(packet.ScriptDebugShapeCircle))
		ps.Colour = protocol.Option(valueOrDefault(shape.Colour, white))
		ps.Location = protocol.Option(vec64To32(shape.Position))
		ps.Scale = protocol.Option(valueOrDefault(float32(shape.Scale), 1))
		ps.Segments = protocol.Option(valueOrDefault(uint8(shape.Segments), 20))
	case *debug.Line:
		ps.Type = protocol.Option(uint8(packet.ScriptDebugShapeLine))
		ps.Colour = protocol.Option(valueOrDefault(shape.Colour, white))
		ps.Location = protocol.Option(vec64To32(shape.Position))
		ps.LineEndLocation = protocol.Option(vec64To32(shape.EndPosition))
	case *debug.Sphere:
		ps.Type = protocol.Option(uint8(packet.ScriptDebugShapeSphere))
		ps.Colour = protocol.Option(valueOrDefault(shape.Colour, white))
		ps.Location = protocol.Option(vec64To32(shape.Position))
		ps.Scale = protocol.Option(valueOrDefault(float32(shape.Scale), 1))
		ps.Segments = protocol.Option(valueOrDefault(uint8(shape.Segments), 20))
	case *debug.Text:
		ps.Type = protocol.Option(uint8(packet.ScriptDebugShapeText))
		ps.Colour = protocol.Option(valueOrDefault(shape.Colour, white))
		ps.Location = protocol.Option(vec64To32(shape.Position))
		ps.Text = protocol.Option(shape.Text)
	default:
		panic(fmt.Sprintf("unknown debug shape type %T", shape))
	}
	return ps
}

// valueOrDefault returns the value passed if it is not the zero value of the type T, otherwise it returns
// the default value provided.
func valueOrDefault[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
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
	return nbtconv.Item(it.NBTData, &s)
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
func stacksToIngredientItems(inputs []recipe.Item) []protocol.ItemDescriptorCount {
	items := make([]protocol.ItemDescriptorCount, 0, len(inputs))
	for _, i := range inputs {
		var d protocol.ItemDescriptor = &protocol.InvalidItemDescriptor{}
		switch i := i.(type) {
		case item.Stack:
			if i.Empty() {
				items = append(items, protocol.ItemDescriptorCount{Descriptor: &protocol.InvalidItemDescriptor{}})
				continue
			}
			rid, meta, ok := world.ItemRuntimeID(i.Item())
			if !ok {
				panic("should never happen")
			}
			if _, ok = i.Value("variants"); ok {
				meta = math.MaxInt16 // Used to indicate that the item has multiple selectable variants.
			}
			d = &protocol.DefaultItemDescriptor{
				NetworkID:     int16(rid),
				MetadataValue: meta,
			}
		case recipe.ItemTag:
			d = &protocol.ItemTagItemDescriptor{Tag: i.Tag()}
		}
		items = append(items, protocol.ItemDescriptorCount{
			Descriptor: d,
			Count:      int32(i.Count()),
		})
	}
	return items
}

// creativeContent returns all creative groups, and creative inventory items as protocol item stacks.
func creativeContent() ([]protocol.CreativeGroup, []protocol.CreativeItem) {
	groups := make([]protocol.CreativeGroup, 0, len(creative.Groups()))
	for _, group := range creative.Groups() {
		groups = append(groups, protocol.CreativeGroup{
			Category: int32(group.Category.Uint8()),
			Name:     group.Name,
			Icon:     deleteDamage(stackFromItem(group.Icon)),
		})
	}

	it := make([]protocol.CreativeItem, 0, len(creative.Items()))
	for index, i := range creative.Items() {
		group := slices.IndexFunc(creative.Groups(), func(group creative.Group) bool {
			return group.Name == i.Group
		})
		if group < 0 {
			continue
		}
		it = append(it, protocol.CreativeItem{
			CreativeItemNetworkID: uint32(index) + 1,
			Item:                  deleteDamage(stackFromItem(i.Stack)),
			GroupIndex:            uint32(group),
		})
	}
	return groups, it
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
	s.FullID = sk.FullID

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

// gameTypeFromMode returns the game type ID from the game mode passed.
func gameTypeFromMode(mode world.GameMode) int32 {
	if mode.AllowsFlying() && mode.CreativeInventory() {
		return packet.GameTypeCreative
	}
	if !mode.Visible() && !mode.HasCollision() {
		return packet.GameTypeSurvivalSpectator
	}
	return packet.GameTypeSurvival
}

// The following functions use the go:linkname directive in order to make sure the item.byID and item.toID
// functions do not need to be exported.

// noinspection ALL
//
//go:linkname item_id github.com/df-mc/dragonfly/server/item.id
func item_id(s item.Stack) int32
