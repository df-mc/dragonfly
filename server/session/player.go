package session

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/recipes"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"go.uber.org/atomic"
	"math"
	"net"
	"time"
	_ "unsafe" // Imported for compiler directives.
)

// closeCurrentContainer closes the container the player might currently have open.
func (s *Session) closeCurrentContainer() {
	if !s.containerOpened.Load() {
		return
	}
	s.closeWindow()
	pos := s.openedPos.Load().(cube.Pos)
	if container, ok := s.c.World().Block(pos).(block.Container); ok {
		container.RemoveViewer(s, s.c.World(), pos)
	}
}

// SendRespawn spawns the controllable of the session client-side in the world, provided it is has died.
func (s *Session) SendRespawn() {
	s.writePacket(&packet.Respawn{
		Position:        vec64To32(s.c.Position().Add(entityOffset(s.c))),
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
	for _, i := range inv.Items() {
		pk.Content = append(pk.Content, instanceFromItem(i))
	}
	s.writePacket(pk)
}

const (
	craftingSizeSmall       = 4
	craftingSizeLarge       = 9
	craftingGridSmallOffset = 28
	craftingGridLargeOffset = 32
	craftingResultIndex     = 50
	craftingFlagAll         = 32767
)

const (
	containerArmour         = 6
	containerChest          = 7
	containerBeacon         = 8
	containerFullInventory  = 12
	containerCraftingGrid   = 13
	containerHotbar         = 27
	containerInventory      = 28
	containerOffHand        = 33
	containerCraftingOffset = 46
	containerBarrel         = 57
	containerCursor         = 58
	containerCreativeOutput = 59
	containerCraftingResult = containerCraftingGrid + containerCraftingOffset
)

// fixID fixes the container ID passed, as it can sometimes be incorrectly sent by Minecraft. (for example, in the recipe book.)
func fixID(id byte) byte {
	if id == containerHotbar || id == containerInventory {
		return containerFullInventory
	}
	return id
}

// invByID attempts to return an inventory by the ID passed. If found, the inventory is returned and the bool
// returned is true.
func (s *Session) invByID(id int32) (*inventory.Inventory, bool) {
	switch id {
	case containerCraftingGrid, containerCreativeOutput, containerCursor:
		// UI inventory.
		return s.ui, true
	case containerHotbar, containerInventory, containerFullInventory:
		// Hotbar 'inventory', rest of inventory, inventory when container is opened.
		return s.inv, true
	case containerOffHand:
		return s.offHand, true
	case containerArmour:
		// Armour inventory.
		return s.armour.Inv(), true
	case containerChest:
		// Chests, potentially other containers too.
		if s.containerOpened.Load() {
			b := s.c.World().Block(s.openedPos.Load().(cube.Pos))
			if _, chest := b.(block.Chest); chest {
				return s.openedWindow.Load().(*inventory.Inventory), true
			}
		}
	case containerBarrel:
		if s.containerOpened.Load() {
			b := s.c.World().Block(s.openedPos.Load().(cube.Pos))
			if _, barrel := b.(block.Barrel); barrel {
				return s.openedWindow.Load().(*inventory.Inventory), true
			}
		}
	case containerBeacon:
		if s.containerOpened.Load() {
			b := s.c.World().Block(s.openedPos.Load().(cube.Pos))
			if _, beacon := b.(block.Beacon); beacon {
				return s.ui, true
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
			Name:    "minecraft:movement",
			Value:   float32(speed),
			Max:     math.MaxFloat32,
			Min:     0,
			Default: 0.1,
		}},
	})
}

// SendCameraShake sends a shake amount for the players camera
func (s *Session) SendCameraShake(Intensity, Duration float32, Type CameraShakeType) {
	s.writePacket(&packet.CameraShake{
		Duration:  Duration,
		Intensity: Intensity,
		Type:      uint8(Type),
	})
}

// SendFood ...
func (s *Session) SendFood(food int, saturation, exhaustion float64) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{
			{
				Name:  "minecraft:player.hunger",
				Value: float32(food),
				Max:   20, Min: 0, Default: 20,
			},
			{
				Name:  "minecraft:player.saturation",
				Value: float32(saturation),
				Max:   20, Min: 0, Default: 20,
			},
			{
				Name:  "minecraft:player.exhaustion",
				Value: float32(exhaustion),
				Max:   5, Min: 0, Default: 0,
			},
		},
	})
}

// SendVelocity sends the velocity of the player to the client.
func (s *Session) SendVelocity(velocity mgl64.Vec3) {
	s.writePacket(&packet.SetActorMotion{
		EntityRuntimeID: selfEntityRuntimeID,
		Velocity:        vec64To32(velocity),
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

// SendGameMode sends the game mode of the Controllable of the session to the client. It makes sure the right
// flags are set to create the full game mode.
func (s *Session) SendGameMode(mode world.GameMode) {
	flags, id, perms := uint32(0), int32(packet.GameTypeSurvivalSpectator), uint32(0)
	if mode.AllowsFlying() {
		flags |= packet.AdventureFlagAllowFlight
	}
	if !mode.HasCollision() {
		flags |= packet.AdventureFlagNoClip
	}
	if !mode.AllowsEditing() {
		flags |= packet.AdventureFlagWorldImmutable
	} else {
		perms |= packet.ActionPermissionBuild | packet.ActionPermissionMine
	}
	if !mode.AllowsInteraction() {
		flags |= packet.AdventureFlagNoPVP
	} else {
		perms |= packet.ActionPermissionDoorsAndSwitched | packet.ActionPermissionOpenContainers | packet.ActionPermissionAttackPlayers | packet.ActionPermissionAttackMobs
	}
	if !mode.Visible() {
		flags |= packet.AdventureFlagMuted
	}
	// Creative or spectator players:
	if mode.AllowsFlying() && mode.CreativeInventory() {
		id = packet.GameTypeCreative
		// Cannot interact with the world, so this is a spectator.
		if !mode.AllowsEditing() && !mode.AllowsInteraction() {
			id = packet.GameTypeCreativeSpectator
		}
	}
	s.writePacket(&packet.AdventureSettings{
		Flags:             flags,
		PermissionLevel:   packet.PermissionLevelMember,
		PlayerUniqueID:    selfEntityRuntimeID,
		ActionPermissions: perms,
	})
	s.writePacket(&packet.SetPlayerGameType{GameType: id})
}

// SendHealth sends the health and max health to the player.
func (s *Session) SendHealth(health *entity.HealthManager) {
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			Name:    "minecraft:health",
			Value:   float32(math.Ceil(health.Health())),
			Max:     float32(math.Ceil(health.MaxHealth())),
			Default: 20,
		}},
	})
}

// SendAbsorption sends the absorption value passed to the player.
func (s *Session) SendAbsorption(value float64) {
	max := value
	if math.Mod(value, 2) != 0 {
		max = value + 1
	}
	s.writePacket(&packet.UpdateAttributes{
		EntityRuntimeID: selfEntityRuntimeID,
		Attributes: []protocol.Attribute{{
			Name:  "minecraft:absorption",
			Value: float32(math.Ceil(value)),
			Max:   float32(math.Ceil(max)),
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

// addToPlayerList adds the player of a session to the player list of this session. It will be shown in the
// in-game pause menu screen.
func (s *Session) addToPlayerList(session *Session) {
	c := session.c

	s.entityMutex.Lock()
	runtimeID := uint64(1)
	if session != s {
		runtimeID = s.currentEntityRuntimeID.Add(1)
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
		FullSkinID:        uuid.New().String(),
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

// HandleInventories starts handling the inventories of the Controllable of the session. It sends packets when
// slots in the inventory are changed.
func (s *Session) HandleInventories() (inv, offHand *inventory.Inventory, armour *inventory.Armour, heldSlot *atomic.Uint32) {
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
			s.writePacket(&packet.InventorySlot{
				WindowID: protocol.WindowIDInventory,
				Slot:     uint32(slot),
				NewItem:  instanceFromItem(item),
			})
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
	s.armour = inventory.NewArmour(func(slot int, item item.Stack) {
		if s.c == nil {
			return
		}
		for _, viewer := range s.c.World().Viewers(s.c.Position()) {
			viewer.ViewEntityArmour(s.c)
		}
		if !s.inTransaction.Load() {
			s.writePacket(&packet.InventorySlot{
				WindowID: protocol.WindowIDArmour,
				Slot:     uint32(slot),
				NewItem:  instanceFromItem(item),
			})
		}
	})
	return s.inv, s.offHand, s.armour, s.heldSlot
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

// stackFromItem converts an item.Stack to its network ItemStack representation.
func stackFromItem(it item.Stack) protocol.ItemStack {
	if it.Empty() {
		return protocol.ItemStack{}
	}
	var blockRuntimeID uint32
	if b, ok := it.Item().(world.Block); ok {
		blockRuntimeID, ok = world.BlockRuntimeID(b)
		if !ok {
			panic("should never happen")
		}
	}

	rid, meta, _ := world.ItemRuntimeID(it.Item())

	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     rid,
			MetadataValue: uint32(meta),
		},
		BlockRuntimeID: int32(blockRuntimeID),
		HasNetworkID:   true,
		Count:          uint16(it.Count()),
		NBTData:        nbtconv.WriteItem(it, false),
	}
}

// instanceFromItem converts an item.Stack to its network ItemInstance representation.
func instanceFromItem(it item.Stack) protocol.ItemInstance {
	return protocol.ItemInstance{
		StackNetworkID: item_id(it),
		Stack:          stackFromItem(it),
	}
}

// stackToItem converts a network ItemStack representation back to an item.Stack.
func stackToItem(it protocol.ItemStack) item.Stack {
	var t world.Item
	var ok bool

	if it.BlockRuntimeID != 0 {
		var b world.Block
		// It shouldn't matter if it (for whatever reason) wasn't able to get the block runtime ID,
		// since on the next line, we assert that the block is an item. If it didn't succeed, it'll
		// return air anyways.
		b, _ = world.BlockByRuntimeID(uint32(it.BlockRuntimeID))
		if t, ok = b.(world.Item); !ok {
			t = block.Air{}
		}
	} else {
		t, ok = world.ItemByRuntimeID(it.NetworkID, int16(it.MetadataValue))
		if !ok {
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

// itemToRecipeIngredientItem converts a recipe.Item into a type that can be used over the protocol.
func itemToRecipeIngredientItem(s recipes.Item) protocol.RecipeIngredientItem {
	if s.Item() == nil {
		return protocol.RecipeIngredientItem{}
	}
	rid, meta, ok := world.ItemRuntimeID(s.Item())
	if !ok {
		panic("should never happen")
	}

	if s.AppliesToAll {
		meta = craftingFlagAll
	}

	return protocol.RecipeIngredientItem{
		NetworkID:     rid,
		MetadataValue: int32(meta),
		Count:         int32(s.Count()),
	}
}

// itemsToRecipeIngredientItems converts a list of recipe.Items into a type that can be used over the protocol.
func itemsToRecipeIngredientItems(s []recipes.Item) (r []protocol.RecipeIngredientItem) {
	for _, st := range s {
		r = append(r, itemToRecipeIngredientItem(st))
	}
	return
}

// protocolRecipes returns all recipes as protocol recipes.
func (s *Session) protocolRecipes() []protocol.Recipe {
	recipeList := make([]protocol.Recipe, 0, len(recipes.All()))
	for index, i := range recipes.All() {
		networkID := uint32(index) + 1
		s.recipeMapping[networkID] = i

		switch newRecipe := i.(type) {
		case recipes.ShapelessRecipe:
			recipeList = append(recipeList, &protocol.ShapelessRecipe{
				RecipeID:        uuid.New().String(),
				Input:           itemsToRecipeIngredientItems(newRecipe.Inputs),
				Output:          []protocol.ItemStack{stackFromItem(newRecipe.Output)},
				Block:           "crafting_table", // TODO: Stop hardcoding this once more blocks that support shapeless recipes are added.
				Priority:        newRecipe.Priority,
				RecipeNetworkID: networkID,
			})
		case recipes.ShapedRecipe:
			recipeList = append(recipeList, &protocol.ShapedRecipe{
				RecipeID:        uuid.New().String(),
				Width:           newRecipe.Dimensions.Width,
				Height:          newRecipe.Dimensions.Height,
				Input:           itemsToRecipeIngredientItems(newRecipe.Inputs),
				Output:          []protocol.ItemStack{stackFromItem(newRecipe.Output)},
				Block:           "crafting_table", // TODO: Stop hardcoding this once more blocks that support shaped recipes are added.
				Priority:        newRecipe.Priority,
				RecipeNetworkID: networkID,
			})
		}
	}
	return recipeList
}

// creativeItems returns all creative inventory items as protocol item stacks.
func creativeItems() []protocol.CreativeItem {
	it := make([]protocol.CreativeItem, 0, len(creative.Items()))
	for index, i := range creative.Items() {
		v := stackFromItem(i)
		delete(v.NBTData, "Damage")
		it = append(it, protocol.CreativeItem{
			CreativeItemNetworkID: uint32(index) + 1,
			Item:                  v,
		})
	}
	return it
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

	m := make(map[string]interface{})
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

// CameraShakeType is the type of camera shake that the player receives
type CameraShakeType uint8

const (
	CameraShakePositional = iota
	CameraShakeRotational
)

// The following functions use the go:linkname directive in order to make sure the item.byID and item.toID
// functions do not need to be exported.

//go:linkname item_id github.com/df-mc/dragonfly/server/item.id
//noinspection ALL
func item_id(s item.Stack) int32
