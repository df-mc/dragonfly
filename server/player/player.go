package player

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player/debug"
	"math"
	"math/rand/v2"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player/bossbar"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/dialogue"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"golang.org/x/text/language"
)

type playerData struct {
	xuid              string
	locale            language.Tag
	nameTag, scoreTag string
	absorptionHealth  float64
	scale             float64

	gameMode world.GameMode
	skin     skin.Skin
	s        *session.Session
	h        Handler

	inv, offHand, enderChest, ui *inventory.Inventory
	armour                       *inventory.Armour
	heldSlot                     *uint32

	sneaking, sprinting, swimming, gliding, crawling, flying,
	invisible, immobile, onGround, usingItem bool
	usingSince time.Time

	glideTicks   int64
	fireTicks    int64
	fallDistance float64

	breathing         bool
	airSupplyTicks    int
	maxAirSupplyTicks int

	cooldowns map[string]time.Time

	speed               float64
	flightSpeed         float64
	verticalFlightSpeed float64

	health     *entity.HealthManager
	experience *entity.ExperienceManager
	effects    *entity.EffectManager

	lastXPPickup *time.Time

	lastDamage  float64
	immuneUntil time.Time

	deathPos       *mgl64.Vec3
	deathDimension world.Dimension

	enchantSeed int64

	mc *entity.MovementComputer

	collidedVertically, collidedHorizontally bool

	breaking          bool
	breakingPos       cube.Pos
	breakingFace      cube.Face
	lastBreakDuration time.Duration

	breakCounter uint32

	hunger *hungerManager

	once sync.Once

	prevWorld *world.World
}

// Player is an implementation of a player entity. It has methods that implement the behaviour that players
// need to play in the world.
type Player struct {
	tx     *world.Tx
	handle *world.EntityHandle
	data   *world.EntityData
	*playerData
}

func (p *Player) H() *world.EntityHandle {
	return p.handle
}

func (p *Player) Tx() *world.Tx {
	return p.tx
}

// Name returns the username of the player. If the player is controlled by a client, it is the username of
// the client. (Typically the XBOX Live name)
func (p *Player) Name() string {
	// TODO: This isn't correct, this will change if the nametag changes.
	return p.data.Name
}

// UUID returns the UUID of the player. This UUID will remain consistent with an XBOX Live account, and will,
// unlike the name of the player, never change.
// It is therefore recommended using the UUID over the name of the player. Additionally, it is recommended to
// use the UUID over the XUID because of its standard format.
func (p *Player) UUID() uuid.UUID {
	return p.handle.UUID()
}

// XUID returns the XBOX Live user ID of the player. It will remain consistent with the XBOX Live account,
// and will not change in the lifetime of an account.
// The XUID is a number that can be parsed as an int64. No more information on what it represents is
// available, and the UUID should be preferred.
// The XUID returned is empty if the Player is not connected to a network session or if the Player is not
// authenticated with XBOX Live.
func (p *Player) XUID() string {
	return p.xuid
}

// DeviceID returns the device ID of the player. If the Player is not connected to a network session, an empty string is
// returned. Otherwise, the device ID the network session sent in the ClientData is returned.
func (p *Player) DeviceID() string {
	if p.session() == session.Nop {
		return ""
	}
	return p.session().ClientData().DeviceID
}

// DeviceModel returns the device model of the player. If the Player is not connected to a network session, an empty
// string is returned. Otherwise, the device model the network session sent in the ClientData is returned.
func (p *Player) DeviceModel() string {
	if p.session() == session.Nop {
		return ""
	}
	return p.session().ClientData().DeviceModel
}

// SelfSignedID returns the self-signed ID of the player. If the Player is not connected to a network session, an empty
// string is returned. Otherwise, the self-signed ID the network session sent in the ClientData is returned.
func (p *Player) SelfSignedID() string {
	if p.session() == session.Nop {
		return ""
	}
	return p.session().ClientData().SelfSignedID
}

// Addr returns the net.Addr of the Player. If the Player is not connected to a network session, nil is returned.
func (p *Player) Addr() net.Addr {
	if p.session() == session.Nop {
		return nil
	}
	return p.session().Addr()
}

// Skin returns the skin that a player is currently using. This skin will be visible to other players
// that the player is shown to.
// If the player was not connected to a network session, a default skin will be set.
func (p *Player) Skin() skin.Skin {
	return p.skin
}

// SetSkin changes the skin of the player. This skin will be visible to other players that the player
// is shown to.
func (p *Player) SetSkin(skin skin.Skin) {
	ctx := event.C(p)
	if p.Handler().HandleSkinChange(ctx, &skin); ctx.Cancelled() {
		p.session().ViewSkin(p)
		return
	}
	p.skin = skin
	for _, v := range p.viewers() {
		v.ViewSkin(p)
	}
}

// Locale returns the language and locale of the Player, as selected in the Player's settings.
func (p *Player) Locale() language.Tag {
	return p.locale
}

// Handle changes the current Handler of the player. As a result, events called by the player will call
// handlers of the Handler passed.
// Handle sets the player's Handler to NopHandler if nil is passed.
func (p *Player) Handle(h Handler) {
	if h == nil {
		h = NopHandler{}
	}
	p.h = h
}

// Message sends a formatted message to the player. The message is formatted following the rules of
// fmt.Sprintln, however the newline at the end is not written.
func (p *Player) Message(a ...any) {
	p.session().SendMessage(format(a))
}

// Messagef sends a formatted message using a specific format to the player. The message is formatted
// according to the fmt.Sprintf formatting rules.
func (p *Player) Messagef(f string, a ...any) {
	p.session().SendMessage(fmt.Sprintf(f, a...))
}

// Messaget sends a translatable message to a player and parameterises it using
// the arguments passed. Messaget panics if an incorrect amount of arguments
// is passed.
func (p *Player) Messaget(t chat.Translation, a ...any) {
	p.session().SendTranslation(t, p.locale, a)
}

// SendPopup sends a formatted popup to the player. The popup is shown above the hotbar of the player and
// overwrites/is overwritten by the name of the item equipped.
// The popup is formatted following the rules of fmt.Sprintln without a newline at the end.
func (p *Player) SendPopup(a ...any) {
	p.session().SendPopup(format(a))
}

// SendTip sends a tip to the player. The tip is shown in the middle of the screen of the player.
// The tip is formatted following the rules of fmt.Sprintln without a newline at the end.
func (p *Player) SendTip(a ...any) {
	p.session().SendTip(format(a))
}

// SendJukeboxPopup sends a formatted jukebox popup to the player. This popup is shown above the hotbar of the player.
// The popup is close to the position of an action bar message and the text has no background.
func (p *Player) SendJukeboxPopup(a ...any) {
	p.session().SendJukeboxPopup(format(a))
}

// SendToast sends a toast to the player. This toast is shown at the top of the screen, similar to achievements or pack
// loading.
func (p *Player) SendToast(title, message string) {
	p.session().SendToast(title, message)
}

// ResetFallDistance resets the player's fall distance.
func (p *Player) ResetFallDistance() {
	p.fallDistance = 0
}

// FallDistance returns the player's fall distance.
func (p *Player) FallDistance() float64 {
	return p.fallDistance
}

// SendTitle sends a title to the player. The title may be configured to change the duration it is displayed
// and the text it shows.
// If non-empty, the subtitle is shown in a smaller font below the title. The same counts for the action text
// of the title, which is shown in a font similar to that of a tip/popup.
func (p *Player) SendTitle(t title.Title) {
	p.session().SetTitleDurations(t.FadeInDuration(), t.Duration(), t.FadeOutDuration())
	if t.Text() != "" || t.Subtitle() != "" {
		p.session().SendTitle(t.Text())
		if t.Subtitle() != "" {
			p.session().SendSubtitle(t.Subtitle())
		}
	}
	if t.ActionText() != "" {
		p.session().SendActionBarMessage(t.ActionText())
	}
}

// SendScoreboard sends a scoreboard to the player. The scoreboard will be present indefinitely until removed
// by the caller.
// SendScoreboard may be called at any time to change the scoreboard of the player.
func (p *Player) SendScoreboard(scoreboard *scoreboard.Scoreboard) {
	p.session().SendScoreboard(scoreboard)
}

// RemoveScoreboard removes any scoreboard currently present on the screen of the player. Nothing happens if
// the player has no scoreboard currently active.
func (p *Player) RemoveScoreboard() {
	p.session().RemoveScoreboard()
}

// SendBossBar sends a boss bar to the player, so that it will be shown indefinitely at the top of the
// player's screen.
// The boss bar may be removed by calling Player.RemoveBossBar().
func (p *Player) SendBossBar(bar bossbar.BossBar) {
	p.session().SendBossBar(bar.Text(), bar.Colour().Uint8(), bar.HealthPercentage())
}

// RemoveBossBar removes any boss bar currently active on the player's screen. If no boss bar is currently
// present, nothing happens.
func (p *Player) RemoveBossBar() {
	p.session().RemoveBossBar()
}

// Chat writes a message in the global chat (chat.Global). The message is prefixed with the name of the
// player and is formatted following the rules of fmt.Sprintln.
func (p *Player) Chat(msg ...any) {
	message := format(msg)
	ctx := event.C(p)
	if p.Handler().HandleChat(ctx, &message); ctx.Cancelled() {
		return
	}
	_, _ = fmt.Fprintf(chat.Global, "<%v> %v\n", p.Name(), message)
}

// ExecuteCommand executes a command passed as the player. If the command could not be found, or if the usage
// was incorrect, an error message is sent to the player. This message should start with a "/" for the command to be
// recognised.
func (p *Player) ExecuteCommand(commandLine string) {
	if p.Dead() {
		return
	}
	args := strings.Split(commandLine, " ")

	name, ok := strings.CutPrefix(args[0], "/")
	if !ok {
		return
	}

	command, ok := cmd.ByAlias(name)
	if !ok {
		o := &cmd.Output{}
		o.Errort(cmd.MessageUnknown, name)
		p.SendCommandOutput(o)
		return
	}
	ctx := event.C(p)
	if p.Handler().HandleCommandExecution(ctx, command, args[1:]); ctx.Cancelled() {
		return
	}
	command.Execute(strings.Join(args[1:], " "), p, p.tx)
}

// Transfer transfers the player to a server at the address passed. If the address could not be resolved, an
// error is returned. If it is returned, the player is closed and transferred to the server.
func (p *Player) Transfer(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	ctx := event.C(p)
	if p.Handler().HandleTransfer(ctx, addr); ctx.Cancelled() {
		return nil
	}
	p.session().Transfer(addr.IP, addr.Port)
	return nil
}

// SendCommandOutput sends the output of a command to the player.
func (p *Player) SendCommandOutput(output *cmd.Output) {
	p.session().SendCommandOutput(output, p.locale)
}

// SendDialogue sends an NPC dialogue to the player, using the entity passed as the entity that the dialogue
// is shown for. Dialogues can be sent on top of each other without the other closing, making it possible
// to have non-flashing transitions between menus compared to forms. The player can either press one of the
// buttons or close the dialogue. It is impossible for a dialogue to have any more than 6 buttons.
func (p *Player) SendDialogue(d dialogue.Dialogue, e world.Entity) {
	p.session().SendDialogue(d, e)
}

// CloseDialogue closes the player's currently open dialogue, if any. If the dialogue's Submittable implements
// dialogue.Closer, the Close method of the Submittable is called after the client acknowledges the closing
// of the dialogue.
func (p *Player) CloseDialogue() {
	p.session().CloseDialogue()
}

// SendForm sends a form to the player for the client to fill out. Once the client fills it out, the Submit
// method of the form will be called.
// Note that the client may also close the form instead of filling it out, which will result in the form not
// having its Submit method called at all. Forms should never depend on the player actually filling out the
// form.
func (p *Player) SendForm(f form.Form) {
	p.session().SendForm(f)
}

// CloseForm closes any forms that the player currently has open. If the player has no forms open, nothing
// happens.
func (p *Player) CloseForm() {
	p.session().CloseForm()
}

// ShowCoordinates enables the vanilla coordinates for the player.
func (p *Player) ShowCoordinates() {
	p.session().EnableCoordinates(true)
}

// HideCoordinates disables the vanilla coordinates for the player.
func (p *Player) HideCoordinates() {
	p.session().EnableCoordinates(false)
}

// EnableInstantRespawn enables the vanilla instant respawn for the player.
func (p *Player) EnableInstantRespawn() {
	p.session().EnableInstantRespawn(true)
}

// DisableInstantRespawn disables the vanilla instant respawn for the player.
func (p *Player) DisableInstantRespawn() {
	p.session().EnableInstantRespawn(false)
}

// SetNameTag changes the name tag displayed over the player in-game. Changing the name tag does not change
// the player's name in, for example, the player list or the chat.
func (p *Player) SetNameTag(name string) {
	p.nameTag = name
	p.updateState()
}

// NameTag returns the current name tag of the Player as shown in-game. It can be changed using SetNameTag.
func (p *Player) NameTag() string {
	return p.nameTag
}

// SetScoreTag changes the score tag displayed over the player in-game. The score tag is displayed under the player's
// name tag.
func (p *Player) SetScoreTag(a ...any) {
	tag := format(a)
	p.scoreTag = tag
	p.updateState()
}

// ScoreTag returns the current score tag of the player. It can be changed using SetScoreTag and by default is empty.
func (p *Player) ScoreTag() string {
	return p.scoreTag
}

// SetSpeed sets the speed of the player. The value passed is the blocks/tick speed that the player will then
// obtain.
func (p *Player) SetSpeed(speed float64) {
	p.speed = speed
	p.session().SendSpeed(speed)
}

// Speed returns the speed of the player, returning a value that indicates the blocks/tick speed. The default
// speed of a player is 0.1.
func (p *Player) Speed() float64 {
	return p.speed
}

// SetFlightSpeed sets the flight speed of the player. The value passed represents the base speed, which is
// multiplied by 10 to obtain the actual blocks/tick speed that the player will then obtain while flying.
func (p *Player) SetFlightSpeed(flightSpeed float64) {
	p.flightSpeed = flightSpeed
	p.session().SendAbilities(p)
}

// FlightSpeed returns the flight speed of the player, with the value representing the base speed. The actual
// blocks/tick speed is this value multiplied by 10. The default flight speed of a player is 0.05, which
// corresponds to 0.5 blocks/tick.
func (p *Player) FlightSpeed() float64 {
	return p.flightSpeed
}

// SetVerticalFlightSpeed sets the flight speed of the player on the Y axis. The value passed represents the
// base speed, which is the blocks/tick speed that the player will obtain while flying.
func (p *Player) SetVerticalFlightSpeed(flightSpeed float64) {
	p.verticalFlightSpeed = flightSpeed
	p.session().SendAbilities(p)
}

// VerticalFlightSpeed returns the flight speed of the player on the Y axis, with the value representing the
// base speed. The default vertical flight speed of a player is 1.0, which corresponds to 1 block/tick.
func (p *Player) VerticalFlightSpeed() float64 {
	return p.verticalFlightSpeed
}

// Health returns the current health of the player. It will always be lower than Player.MaxHealth().
func (p *Player) Health() float64 {
	return p.health.Health()
}

// MaxHealth returns the maximum amount of health that a player may have. The MaxHealth will always be higher
// than Player.Health().
func (p *Player) MaxHealth() float64 {
	return p.health.MaxHealth()
}

// SetMaxHealth sets the maximum health of the player. If the current health of the player is higher than the
// new maximum health, the health is set to the new maximum.
// SetMaxHealth panics if the max health passed is 0 or lower.
func (p *Player) SetMaxHealth(health float64) {
	p.health.SetMaxHealth(health)
	p.session().SendHealth(p.Health(), p.MaxHealth(), p.absorptionHealth)
}

// addHealth adds health to the player's current health.
func (p *Player) addHealth(health float64) {
	p.health.AddHealth(health)
	p.session().SendHealth(p.Health(), p.MaxHealth(), p.absorptionHealth)
}

// Heal heals the entity for a given amount of health. The source passed
// represents the cause of the healing, for example entity.FoodHealingSource if
// the entity healed by having a full food bar. If the health added to the
// original health exceeds the entity's max health, Heal will not add the full
// amount. If the health passed is negative, Heal will not do anything.
func (p *Player) Heal(health float64, source world.HealingSource) {
	if p.Dead() || health < 0 || !p.GameMode().AllowsTakingDamage() {
		return
	}
	ctx := event.C(p)
	if p.Handler().HandleHeal(ctx, &health, source); ctx.Cancelled() {
		return
	}
	p.addHealth(health)
}

// updateFallState is called to update the entities falling state.
func (p *Player) updateFallState(distanceThisTick float64) {
	if p.OnGround() {
		if p.fallDistance > 0 {
			p.fall(p.fallDistance)
			p.ResetFallDistance()
		}
	} else if distanceThisTick < p.fallDistance {
		p.fallDistance -= distanceThisTick
	} else {
		p.ResetFallDistance()
	}
}

// fall is called when a falling entity hits the ground.
func (p *Player) fall(distance float64) {
	pos := cube.PosFromVec3(p.Position())
	b := p.tx.Block(pos)

	if len(b.Model().BBox(pos, p.tx)) == 0 {
		pos = pos.Sub(cube.Pos{0, 1})
		b = p.tx.Block(pos)
	}
	if h, ok := b.(block.EntityLander); ok {
		h.EntityLand(pos, p.tx, p, &distance)
	}
	dmg := distance - 3
	if boost, ok := p.Effect(effect.JumpBoost); ok {
		dmg -= float64(boost.Level())
	}
	if dmg < 0.5 {
		return
	}
	p.Hurt(math.Ceil(dmg), entity.FallDamageSource{})
}

// Hurt hurts the player for a given amount of damage. The source passed
// represents the cause of the damage, for example entity.AttackDamageSource if
// the player is attacked by another entity. If the final damage exceeds the
// health that the player currently has, the player is killed and will have to
// respawn.
// If the damage passed is negative, Hurt will not do anything. Hurt returns the
// final damage dealt to the Player and if the Player was vulnerable to this
// kind of damage.
func (p *Player) Hurt(dmg float64, src world.DamageSource) (float64, bool) {
	if _, ok := p.Effect(effect.FireResistance); (ok && src.Fire()) || p.Dead() || !p.GameMode().AllowsTakingDamage() || dmg < 0 {
		return 0, false
	}
	totalDamage := p.FinalDamageFrom(dmg, src)
	damageLeft := totalDamage

	immune := time.Now().Before(p.immuneUntil)
	if immune {
		if damageLeft = damageLeft - p.lastDamage; damageLeft <= 0 {
			return 0, false
		}
	}

	immunity := time.Second / 2
	ctx := event.C(p)
	if p.Handler().HandleHurt(ctx, &damageLeft, immune, &immunity, src); ctx.Cancelled() {
		return 0, false
	}
	p.setAttackImmunity(immunity, totalDamage)

	if a := p.Absorption(); a > 0 {
		p.SetAbsorption(a - damageLeft)
		damageLeft = max(0, damageLeft-a)
	}

	if p.Health()-damageLeft <= mgl64.Epsilon && !src.IgnoreTotem() {
		hand, offHand := p.HeldItems()
		if _, ok := offHand.Item().(item.Totem); ok {
			p.applyTotemEffects()
			p.SetHeldItems(hand, offHand.Grow(-1))
			return 0, false
		} else if _, ok := hand.Item().(item.Totem); ok {
			p.applyTotemEffects()
			p.SetHeldItems(hand.Grow(-1), offHand)
			return 0, false
		}
	}

	p.addHealth(-damageLeft)

	if src.ReducedByArmour() {
		p.Exhaust(0.1)
		p.Armour().Damage(dmg, p.damageItem)

		var origin world.Entity
		if s, ok := src.(entity.AttackDamageSource); ok {
			origin = s.Attacker
		} else if s, ok := src.(entity.ProjectileDamageSource); ok {
			origin = s.Owner
		}
		if l, ok := origin.(entity.Living); ok {
			if thornsDmg := p.Armour().ThornsDamage(p.damageItem); thornsDmg > 0 {
				l.Hurt(thornsDmg, enchantment.ThornsDamageSource{Owner: p})
			}
		}
	}

	pos := p.Position()
	for _, viewer := range p.viewers() {
		viewer.ViewEntityAction(p, entity.HurtAction{})
	}
	if src.Fire() {
		p.tx.PlaySound(pos, sound.Burning{})
	} else if _, ok := src.(entity.DrowningDamageSource); ok {
		p.tx.PlaySound(pos, sound.Drowning{})
	}

	if p.Dead() {
		p.kill(src)
	}
	return totalDamage, true
}

// applyTotemEffects is an unexported function that is used to handle totem effects.
func (p *Player) applyTotemEffects() {
	p.addHealth(2 - p.Health())

	for _, e := range p.Effects() {
		p.RemoveEffect(e.Type())
	}

	p.AddEffect(effect.New(effect.Regeneration, 2, time.Second*40))
	p.AddEffect(effect.New(effect.FireResistance, 1, time.Second*40))
	p.AddEffect(effect.New(effect.Absorption, 2, time.Second*5))

	p.tx.PlaySound(p.Position(), sound.Totem{})

	for _, viewer := range p.viewers() {
		viewer.ViewEntityAction(p, entity.TotemUseAction{})
	}
}

// FinalDamageFrom resolves the final damage received by the player if it is attacked by the source passed
// with the damage passed. FinalDamageFrom takes into account things such as the armour worn and the
// enchantments on the individual pieces.
// The damage returned will be at the least 0.
func (p *Player) FinalDamageFrom(dmg float64, src world.DamageSource) float64 {
	dmg = max(dmg, 0)

	dmg -= p.Armour().DamageReduction(dmg, src)
	if res, ok := p.Effect(effect.Resistance); ok {
		dmg *= effect.Resistance.Multiplier(src, res.Level())
	}
	return dmg
}

// Explode ...
func (p *Player) Explode(explosionPos mgl64.Vec3, impact float64, c block.ExplosionConfig) {
	diff := p.Position().Sub(explosionPos)
	p.Hurt(math.Floor((impact*impact+impact)*3.5*c.Size*2+1), entity.ExplosionDamageSource{})
	p.knockBack(explosionPos, impact, diff[1]/diff.Len()*impact)
}

// SetAbsorption sets the absorption health of a player. This extra health shows as golden hearts and do not
// actually increase the maximum health. Once the hearts are lost, they will not regenerate.
// Nothing happens if a negative number is passed.
func (p *Player) SetAbsorption(health float64) {
	p.absorptionHealth = max(health, 0)
	p.session().SendHealth(p.Health(), p.MaxHealth(), p.absorptionHealth)
}

// Absorption returns the absorption health that the player has.
func (p *Player) Absorption() float64 {
	return p.absorptionHealth
}

// KnockBack knocks the player back with a given force and height. A source is passed which indicates the
// source of the velocity, typically the position of an attacking entity. The source is used to calculate the
// direction which the entity should be knocked back in.
func (p *Player) KnockBack(src mgl64.Vec3, force, height float64) {
	if p.Dead() || !p.GameMode().AllowsTakingDamage() {
		return
	}
	p.knockBack(src, force, height)
}

// knockBack is an unexported function that is used to knock the player back. This function does not check if the player
// can take damage or not.
func (p *Player) knockBack(src mgl64.Vec3, force, height float64) {
	velocity := p.Position().Sub(src)
	velocity[1] = 0

	if velocity.Len() != 0 {
		velocity = velocity.Normalize().Mul(force)
	}
	velocity[1] = height

	p.SetVelocity(velocity.Mul(1 - p.Armour().KnockBackResistance()))
}

// setAttackImmunity sets the duration the player is immune to entity attacks.
func (p *Player) setAttackImmunity(d time.Duration, dmg float64) {
	p.immuneUntil = time.Now().Add(d)
	p.lastDamage = dmg
}

// Food returns the current food level of a player. The level returned is guaranteed to always be between 0
// and 20. Every half drumstick is one level.
func (p *Player) Food() int {
	return p.hunger.Food()
}

// SetFood sets the food level of a player. The level passed must be in a range of 0-20. If the level passed
// is negative, the food level will be set to 0. If the level exceeds 20, the food level will be set to 20.
func (p *Player) SetFood(level int) {
	p.hunger.SetFood(level)
	p.sendFood()
}

// AddFood adds a number of points to the food level of the player. If the new food level is negative or if
// it exceeds 20, it will be set to 0 or 20 respectively.
func (p *Player) AddFood(points int) {
	p.hunger.AddFood(points)
	p.sendFood()
}

// Saturate saturates the player's food bar with the amount of food points and saturation points passed. The
// total saturation of the player will never exceed its total food level.
func (p *Player) Saturate(food int, saturation float64) {
	p.hunger.saturate(food, saturation)
	p.sendFood()
}

// sendFood sends the current food properties to the client.
func (p *Player) sendFood() {
	p.session().SendFood(p.hunger.foodLevel, p.hunger.saturationLevel, p.hunger.exhaustionLevel)
}

// AddEffect adds an entity.Effect to the Player. If the effect is instant, it is applied to the Player
// immediately. If not, the effect is applied to the player every time the Tick method is called.
// AddEffect will overwrite any effects present if the level of the effect is higher than the existing one, or
// if the effects' levels are equal and the new effect has a longer duration.
func (p *Player) AddEffect(e effect.Effect) {
	p.session().SendEffect(p.effects.Add(e, p))
	p.updateState()
}

// RemoveEffect removes any effect that might currently be active on the Player.
func (p *Player) RemoveEffect(e effect.Type) {
	p.effects.Remove(e, p)
	p.session().SendEffectRemoval(e)
	p.updateState()
}

// Effect returns the effect instance and true if the Player has the effect. If not found, it will return an empty
// effect instance and false.
func (p *Player) Effect(e effect.Type) (effect.Effect, bool) {
	return p.effects.Effect(e)
}

// Effects returns any effect currently applied to the entity. The returned effects are guaranteed not to have
// expired when returned.
func (p *Player) Effects() []effect.Effect {
	return p.effects.Effects()
}

// BeaconAffected ...
func (*Player) BeaconAffected() bool {
	return true
}

// Exhaust exhausts the player by the amount of points passed if the player is in survival mode. If the total
// exhaustion level exceeds 4, a saturation point, or food point, if saturation is 0, will be subtracted.
func (p *Player) Exhaust(points float64) {
	if !p.GameMode().AllowsTakingDamage() || p.tx.World().Difficulty().FoodRegenerates() {
		return
	}
	before := p.hunger.Food()
	p.hunger.exhaust(points)
	if after := p.hunger.Food(); before != after {
		// Temporarily set the food level back so that it hasn't yet changed once the event is handled.
		p.hunger.SetFood(before)

		ctx := event.C(p)
		if p.Handler().HandleFoodLoss(ctx, before, &after); ctx.Cancelled() {
			// Reset the exhaustion level if the event was cancelled. Because if
			// we cancel this, and at some point we stop cancelling it, the
			// first food point will be lost more quickly than expected.
			p.hunger.resetExhaustion()
			return
		}
		p.hunger.SetFood(after)
		if before >= 7 && after <= 6 {
			// The client will stop sprinting by itself too, but we force it just to be sure.
			p.StopSprinting()
		}
	}
	p.sendFood()
}

// Dead checks if the player is considered dead. True is returned if the health of the player is equal to or
// lower than 0.
func (p *Player) Dead() bool {
	return p.Health() <= mgl64.Epsilon
}

// DeathPosition returns the last position the player was at when they died. If the player has never died, the third
// return value will be false.
func (p *Player) DeathPosition() (mgl64.Vec3, world.Dimension, bool) {
	if p.deathPos == nil {
		return mgl64.Vec3{}, nil, false
	}
	return *p.deathPos, p.deathDimension, true
}

// kill kills the player, clearing its inventories and resetting it to its base state.
func (p *Player) kill(src world.DamageSource) {
	for _, viewer := range p.viewers() {
		viewer.ViewEntityAction(p, entity.DeathAction{})
	}

	p.addHealth(-p.MaxHealth())

	keepInv := false
	p.Handler().HandleDeath(p, src, &keepInv)
	p.StopSneaking()
	p.StopSprinting()

	pos := p.Position()
	if !keepInv {
		p.dropItems()
	}
	for _, e := range p.Effects() {
		p.RemoveEffect(e.Type())
	}

	p.deathPos, p.deathDimension = &pos, p.tx.World().Dimension()

	// Wait a little before removing the entity. The client displays a death
	// animation while the player is dying.
	time.AfterFunc(time.Millisecond*1100, func() {
		p.H().ExecWorld(finishDying)
	})
}

// finishDying completes the death of a player, removing it from the world.
func finishDying(_ *world.Tx, e world.Entity) {
	p := e.(*Player)
	if p.session() == session.Nop {
		_ = p.Close()
		return
	}
	if p.Dead() {
		p.SetInvisible()
		// We have an actual client connected to this player: We change its
		// position server side so that in the future, the client won't respawn
		// on the death location when disconnecting. The client should not see
		// the movement itself yet, though.
		p.data.Pos = p.tx.World().Spawn().Vec3()
	}
}

// dropItems drops all items and experience of the Player on the ground in random directions.
func (p *Player) dropItems() {
	pos := p.Position()
	for _, orb := range entity.NewExperienceOrbs(pos, int(math.Min(float64(p.experience.Level()*7), 100))) {
		p.tx.AddEntity(orb)
	}
	p.experience.Reset()
	p.session().SendExperience(p.ExperienceLevel(), p.ExperienceProgress())

	p.MoveItemsToInventory()
	for _, it := range append(p.inv.Clear(), append(p.armour.Clear(), p.offHand.Clear()...)...) {
		if _, ok := it.Enchantment(enchantment.CurseOfVanishing); ok {
			continue
		}
		opts := world.EntitySpawnOpts{Position: pos, Velocity: mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1}}
		p.tx.AddEntity(entity.NewItem(opts, it))
	}
}

// MoveItemsToInventory moves items kept in 'temporary' slots, such as the
// crafting grid of slots in an enchantment table, to the player's inventory.
// If no space is left for these items, the leftover items are dropped.
func (p *Player) MoveItemsToInventory() {
	for _, i := range p.ui.Clear() {
		if n, err := p.inv.AddItem(i); err != nil {
			// We couldn't add the item to the main inventory (probably because
			// it was full), so we drop it instead.
			p.Drop(i.Grow(i.Count() - n))
		}
	}
}

// Respawn spawns the player after it dies, so that its health is replenished,
// and it is spawned in the world again. Nothing will happen if the player does
// not have a session connected to it.
// Calling Respawn may lead to the player being removed from its world and being
// added to a new world. This means that p cannot be assumed to be valid after
// a call to Respawn.
func (p *Player) Respawn() *world.EntityHandle {
	p.respawn(nil)
	return p.handle
}

func (p *Player) respawn(f func(p *Player)) {
	if !p.Dead() || p.session() == session.Nop {
		return
	}
	// We can use the principle here that returning through a portal of a specific dimension inside that dimension will
	// always bring us back to the overworld.
	w := p.tx.World().PortalDestination(p.tx.World().Dimension())
	pos := w.PlayerSpawn(p.UUID()).Vec3Middle()

	p.addHealth(p.MaxHealth())
	p.hunger.Reset()
	p.sendFood()
	p.Extinguish()
	p.ResetFallDistance()

	p.Handler().HandleRespawn(p, &pos, &w)

	handle := p.tx.RemoveEntity(p)
	w.Exec(func(tx *world.Tx) {
		np := tx.AddEntity(handle).(*Player)
		np.Teleport(pos)
		np.session().SendRespawn(pos, p)
		np.SetVisible()
		if f != nil {
			f(np)
		}
	})
}

// StartSprinting makes a player start sprinting, increasing the speed of the player by 30% and making
// particles show up under the feet. The player will only start sprinting if its food level is high enough.
// If the player is sneaking when calling StartSprinting, it is stopped from sneaking.
func (p *Player) StartSprinting() {
	if !p.hunger.canSprint() && p.GameMode().AllowsTakingDamage() || p.crawling || p.sprinting {
		return
	}
	ctx := event.C(p)
	if p.Handler().HandleToggleSprint(ctx, true); ctx.Cancelled() {
		return
	}
	p.StopSneaking()
	p.sprinting = true
	p.SetSpeed(p.speed * 1.3)
	p.updateState()
}

// Sprinting checks if the player is currently sprinting.
func (p *Player) Sprinting() bool {
	return p.sprinting
}

// StopSprinting makes a player stop sprinting, setting back the speed of the player to its original value.
func (p *Player) StopSprinting() {
	if !p.sprinting {
		return
	}
	ctx := event.C(p)
	if p.Handler().HandleToggleSprint(ctx, false); ctx.Cancelled() {
		return
	}
	p.sprinting = false
	p.SetSpeed(p.speed / 1.3)
	p.updateState()
}

// StartSneaking makes a player start sneaking. If the player is already sneaking, StartSneaking will not do
// anything.
// If the player is sprinting while StartSneaking is called, the sprinting is stopped.
func (p *Player) StartSneaking() {
	if p.sneaking {
		return
	}
	ctx := event.C(p)
	if p.Handler().HandleToggleSneak(ctx, true); ctx.Cancelled() {
		return
	}
	if !p.Flying() {
		p.StopSprinting()
	}
	p.sneaking = true
	p.updateState()
}

// Sneaking checks if the player is currently sneaking.
func (p *Player) Sneaking() bool {
	return p.sneaking
}

// StopSneaking makes a player stop sneaking if it currently is. If the player is not sneaking, StopSneaking
// will not do anything.
func (p *Player) StopSneaking() {
	if !p.sneaking {
		return
	}
	ctx := event.C(p)
	if p.Handler().HandleToggleSneak(ctx, false); ctx.Cancelled() {
		return
	}
	p.sneaking = false
	p.updateState()
}

// StartSwimming makes the player start swimming if it is not currently doing so. If the player is sneaking
// while StartSwimming is called, the sneaking is stopped.
func (p *Player) StartSwimming() {
	if p.swimming {
		return
	}
	p.StopSneaking()
	p.swimming = true
	p.updateState()
}

// Swimming checks if the player is currently swimming.
func (p *Player) Swimming() bool {
	return p.swimming
}

// StopSwimming makes the player stop swimming if it is currently doing so.
func (p *Player) StopSwimming() {
	if !p.swimming {
		return
	}
	p.swimming = false
	p.updateState()
}

// Splash is called when a water bottle splashes onto the player.
func (p *Player) Splash(*world.Tx, mgl64.Vec3) {
	if d := p.OnFireDuration(); d.Seconds() <= 0 {
		return
	}
	p.Extinguish()
}

// StartCrawling makes the player start crawling if it is not currently doing so. If the player is sneaking
// while StartCrawling is called, the sneaking is stopped.
func (p *Player) StartCrawling() {
	if p.crawling {
		return
	}
	for _, corner := range p.H().Type().BBox(p).Translate(p.Position()).Corners() {
		if _, isAir := p.tx.Block(cube.PosFromVec3(corner).Add(cube.Pos{0, 1, 0})).(block.Air); !isAir {
			p.crawling = true
			break
		}
	}
	p.StopSneaking()
	p.updateState()
}

// StopCrawling makes the player stop crawling if it is currently doing so.
func (p *Player) StopCrawling() {
	if !p.crawling {
		return
	}
	p.crawling = false
	p.updateState()
}

// Crawling checks if the player is currently crawling.
func (p *Player) Crawling() bool {
	return p.crawling
}

// StartGliding makes the player start gliding if it is not currently doing so.
func (p *Player) StartGliding() {
	if p.gliding {
		return
	}
	chest := p.Armour().Chestplate()
	if _, ok := chest.Item().(item.Elytra); !ok || chest.Durability() < 2 {
		return
	}
	p.gliding = true
	p.updateState()
}

// Gliding checks if the player is currently gliding.
func (p *Player) Gliding() bool {
	return p.gliding
}

// StopGliding makes the player stop gliding if it is currently doing so.
func (p *Player) StopGliding() {
	if !p.gliding {
		return
	}
	p.gliding = false
	p.glideTicks = 0
	p.updateState()
}

// StartFlying makes the player start flying if they aren't already. It requires the player to be in a gamemode which
// allows flying.
func (p *Player) StartFlying() {
	if !p.GameMode().AllowsFlying() || p.Flying() {
		return
	}
	p.flying = true
	p.session().SendGameMode(p)
}

// Flying checks if the player is currently flying.
func (p *Player) Flying() bool {
	return p.flying
}

// StopFlying makes the player stop flying if it currently is.
func (p *Player) StopFlying() {
	if !p.flying {
		return
	}
	p.flying = false
	p.session().SendGameMode(p)
}

// Jump makes the player jump if they are on ground. It exhausts the player by 0.05 food points, an additional 0.15
// is exhausted if the player is sprint jumping.
func (p *Player) Jump() {
	if p.Dead() {
		return
	}

	p.Handler().HandleJump(p)
	if p.OnGround() {
		jumpVel := 0.42
		if e, ok := p.Effect(effect.JumpBoost); ok {
			jumpVel = float64(e.Level()) / 10
		}
		p.data.Vel = mgl64.Vec3{0, jumpVel}
	}
	if p.Sprinting() {
		p.Exhaust(0.2)
	} else {
		p.Exhaust(0.05)
	}
}

// SetInvisible sets the player invisible, so that other players will not be able to see it.
func (p *Player) SetInvisible() {
	if p.Invisible() {
		return
	}
	p.invisible = true
	p.updateState()
}

// SetVisible sets the player visible again, so that other players can see it again. If the player was already
// visible, or if the player is in spectator mode, nothing happens.
func (p *Player) SetVisible() {
	if _, ok := p.Effect(effect.Invisibility); ok || !p.GameMode().Visible() || !p.invisible {
		return
	}
	p.invisible = false
	p.updateState()
}

// Invisible checks if the Player is currently invisible.
func (p *Player) Invisible() bool {
	return p.invisible
}

// SetImmobile prevents the player from moving around, but still allows them to look around.
func (p *Player) SetImmobile() {
	if p.Immobile() {
		return
	}
	p.immobile = true
	p.updateState()
}

// SetMobile allows the player to freely move around again after being immobile.
func (p *Player) SetMobile() {
	if !p.Immobile() {
		return
	}
	p.immobile = false
	p.updateState()
}

// Immobile checks if the Player is currently immobile.
func (p *Player) Immobile() bool {
	return p.immobile
}

// FireProof checks if the Player is currently fireproof. True is returned if the player has a fireResistance effect or
// if it is in creative mode.
func (p *Player) FireProof() bool {
	if _, ok := p.Effect(effect.FireResistance); ok {
		return true
	}
	return !p.GameMode().AllowsTakingDamage()
}

// OnFireDuration ...
func (p *Player) OnFireDuration() time.Duration {
	return time.Duration(p.fireTicks) * time.Second / 20
}

// SetOnFire ...
func (p *Player) SetOnFire(duration time.Duration) {
	ticks := int64(duration.Seconds() * 20)
	if level := p.Armour().HighestEnchantmentLevel(enchantment.FireProtection); level > 0 {
		ticks -= int64(math.Floor(float64(ticks) * float64(level) * 0.15))
	}
	p.fireTicks = ticks
	p.updateState()
}

// Extinguish ...
func (p *Player) Extinguish() {
	p.SetOnFire(0)
}

// Inventory returns the inventory of the player. This inventory holds the items stored in the normal part of
// the inventory and the hotbar. It also includes the item in the main hand as returned by Player.HeldItems().
func (p *Player) Inventory() *inventory.Inventory {
	return p.inv
}

// Armour returns the armour inventory of the player. This inventory yields 4 slots, for the helmet,
// chestplate, leggings and boots respectively.
func (p *Player) Armour() *inventory.Armour {
	return p.armour
}

// HeldItems returns the items currently held in the hands of the player. The first item stack returned is the
// one held in the main hand, the second is held in the off-hand.
// If no item was held in a hand, the stack returned has a count of 0. Stack.Empty() may be used to check if
// the hand held anything.
func (p *Player) HeldItems() (mainHand, offHand item.Stack) {
	offHand, _ = p.offHand.Item(0)
	mainHand, _ = p.inv.Item(int(*p.heldSlot))
	return mainHand, offHand
}

// SetHeldItems sets items to the main hand and the off-hand of the player. The Stacks passed may be empty
// (Stack.Empty()) to clear the held item.
func (p *Player) SetHeldItems(mainHand, offHand item.Stack) {
	_ = p.inv.SetItem(int(*p.heldSlot), mainHand)
	_ = p.offHand.SetItem(0, offHand)
}

// SetHeldSlot updates the held slot of the player to the slot provided. The
// slot must be between 0 and 8.
func (p *Player) SetHeldSlot(to int) error {
	// The slot that the player might have selected must be within the hotbar:
	// The held item cannot be in a different place in the inventory.
	if to < 0 || to > 8 {
		return fmt.Errorf("held slot exceeds hotbar range 0-8: slot is %v", to)
	}
	from := int(*p.heldSlot)
	if from == to {
		// Old slot was the same as new slot, so don't do anything.
		return nil
	}

	ctx := event.C(p)
	p.Handler().HandleHeldSlotChange(ctx, from, to)
	if ctx.Cancelled() {
		// The slot change was cancelled, resend held slot.
		p.session().SendHeldSlot(from, p, true)
		return nil
	}
	*p.heldSlot = uint32(to)
	p.usingItem = false

	for _, viewer := range p.viewers() {
		viewer.ViewEntityItems(p)
	}
	p.session().SendHeldSlot(to, p, false)
	return nil
}

// EnderChestInventory returns the player's ender chest inventory. Its accessed by the player when opening
// ender chests anywhere.
func (p *Player) EnderChestInventory() *inventory.Inventory {
	return p.enderChest
}

// SetGameMode sets the game mode of a player. The game mode specifies the way that the player can interact
// with the world that it is in.
func (p *Player) SetGameMode(mode world.GameMode) {
	previous := p.GameMode()
	p.gameMode = mode

	if !mode.AllowsFlying() {
		p.StopFlying()
	}
	if !mode.Visible() {
		p.SetInvisible()
	} else if !previous.Visible() {
		p.SetVisible()
	}

	p.session().SendGameMode(p)
	for _, v := range p.viewers() {
		v.ViewEntityGameMode(p)
	}
	if mode.AllowsTakingDamage() {
		p.session().SendHealth(p.Health(), p.MaxHealth(), p.absorptionHealth)
	}
}

// GameMode returns the current game mode assigned to the player. If not changed, the game mode returned will
// be the same as that of the world that the player spawns in.
// The game mode may be changed using Player.SetGameMode().
func (p *Player) GameMode() world.GameMode {
	return p.gameMode
}

// HasCooldown returns true if the item passed has an active cooldown, meaning it currently cannot be used again. If the
// world.Item passed is nil, HasCooldown always returns false.
func (p *Player) HasCooldown(item world.Item) bool {
	if item == nil {
		return false
	}
	name, _ := item.EncodeItem()
	otherTime, ok := p.cooldowns[name]
	if !ok {
		return false
	}
	if time.Now().After(otherTime) {
		delete(p.cooldowns, name)
		return false
	}
	return true
}

// SetCooldown sets a cooldown for an item. If the world.Item passed is nil, nothing happens.
func (p *Player) SetCooldown(item world.Item, cooldown time.Duration) {
	if item == nil {
		return
	}
	name, _ := item.EncodeItem()
	p.cooldowns[name] = time.Now().Add(cooldown)
	p.session().ViewItemCooldown(item, cooldown)
}

// UseItem uses the item currently held in the player's main hand in the air. Generally, nothing happens,
// unless the held item implements the item.Usable interface, in which case it will be activated.
// This generally happens for items such as throwable items like snowballs.
func (p *Player) UseItem() {
	var (
		i, left = p.HeldItems()
		ctx     = event.C(p)
	)
	if p.HasCooldown(i.Item()) {
		return
	}
	if p.Handler().HandleItemUse(ctx); ctx.Cancelled() {
		return
	}
	i, left = p.HeldItems()
	it := i.Item()

	if cd, ok := it.(item.Cooldown); ok {
		p.SetCooldown(it, cd.Cooldown())
	}

	if _, ok := it.(item.Releasable); ok {
		if !p.canRelease() {
			return
		}
		p.usingSince, p.usingItem = time.Now(), true
		p.updateState()
	}
	switch usable := it.(type) {
	case item.Chargeable:
		useCtx := p.useContext()
		if !p.usingItem {
			if !usable.ReleaseCharge(p, p.tx, useCtx) {
				// If the item was not charged yet, start charging.
				p.usingSince, p.usingItem = time.Now(), true
			}
			p.handleUseContext(useCtx)
			p.updateState()
			return
		}

		// Stop charging and determine if the item is ready.
		p.usingItem = false
		dur := p.useDuration()
		if usable.Charge(p, p.tx, useCtx, dur) {
			p.session().SendChargeItemComplete()
		}
		p.handleUseContext(useCtx)
		p.updateState()
	case item.Usable:
		useCtx := p.useContext()
		if !usable.Use(p.tx, p, useCtx) {
			return
		}
		// We only swing the player's arm if the item held actually does something. If it doesn't, there is no
		// reason to swing the arm.
		p.SwingArm()
		p.SetHeldItems(p.subtractItem(p.damageItem(i, useCtx.Damage), useCtx.CountSub), left)
		p.addNewItem(useCtx)
	case item.Consumable:
		if c, ok := usable.(interface{ CanConsume() bool }); ok && !c.CanConsume() {
			p.ReleaseItem()
			return
		}
		if !usable.AlwaysConsumable() && p.GameMode().AllowsTakingDamage() && p.Food() >= 20 {
			// The item.Consumable is not always consumable, the player is not in creative mode and the
			// food bar is filled: The item cannot be consumed.
			p.ReleaseItem()
			return
		}
		if !p.usingItem {
			// Consumable starts being consumed: Set the start timestamp and update the using state to viewers.
			p.usingItem, p.usingSince = true, time.Now()
			p.updateState()
			return
		}
		// The player is currently using the item held. This is a signal the item was consumed, so we
		// consume it and start using it again.
		useCtx, dur := p.useContext(), p.useDuration()
		if dur < usable.ConsumeDuration() {
			// The required duration for consuming this item was not met, so we don't consume it.
			return
		}
		// Reset the duration for the next item to be consumed.
		p.usingSince = time.Now()
		ctx := event.C(p)
		if p.Handler().HandleItemConsume(ctx, i); ctx.Cancelled() {
			return
		}
		useCtx.CountSub, useCtx.NewItem = 1, usable.Consume(p.tx, p)
		p.handleUseContext(useCtx)
		p.tx.PlaySound(p.Position().Add(mgl64.Vec3{0, 1.5}), sound.Burp{})
	}
}

// ReleaseItem makes the Player release the item it is currently using. This is only applicable for items that
// implement the item.Releasable interface.
// If the Player is not currently using any item, ReleaseItem returns immediately.
// ReleaseItem either aborts the using of the item or finished it, depending on the time that elapsed since
// the item started being used.
func (p *Player) ReleaseItem() {
	if !p.usingItem || !p.canRelease() || !p.GameMode().AllowsInteraction() {
		p.usingItem = false
		return
	}
	p.usingItem = false

	useCtx, dur := p.useContext(), p.useDuration()
	i, _ := p.HeldItems()
	ctx := event.C(p)
	if p.Handler().HandleItemRelease(ctx, i, dur); ctx.Cancelled() {
		return
	}
	i.Item().(item.Releasable).Release(p, p.tx, useCtx, dur)
	p.handleUseContext(useCtx)
	p.updateState()
}

// canRelease returns whether the player can release the item currently held in the main hand.
func (p *Player) canRelease() bool {
	held, left := p.HeldItems()
	releasable, ok := held.Item().(item.Releasable)
	if !ok {
		return false
	}
	if p.GameMode().CreativeInventory() {
		return true
	}
	for _, req := range releasable.Requirements() {
		reqName, _ := req.Item().EncodeItem()

		if !left.Empty() {
			leftName, _ := left.Item().EncodeItem()
			if leftName == reqName {
				continue
			}
		}

		_, found := p.Inventory().FirstFunc(func(stack item.Stack) bool {
			name, _ := stack.Item().EncodeItem()
			return name == reqName
		})
		if !found {
			return false
		}
	}
	return true
}

// handleUseContext handles the item.UseContext after the item has been used.
func (p *Player) handleUseContext(ctx *item.UseContext) {
	i, left := p.HeldItems()

	p.SetHeldItems(p.subtractItem(p.damageItem(i, ctx.Damage), ctx.CountSub), left)
	p.addNewItem(ctx)
	for _, it := range ctx.ConsumedItems {
		_, offHand := p.HeldItems()
		if offHand.Comparable(it) {
			if err := p.offHand.RemoveItem(it); err == nil {
				continue
			}

			it = it.Grow(-offHand.Count())
		}

		_ = p.Inventory().RemoveItem(it)
	}
}

// useDuration returns the duration the player has been using the item in the main hand.
func (p *Player) useDuration() time.Duration {
	return time.Since(p.usingSince) + time.Second/20
}

// UsingItem checks if the Player is currently using an item. True is returned if the Player is currently eating an
// item or using it over a longer duration such as when using a bow.
func (p *Player) UsingItem() bool {
	return p.usingItem
}

// UseItemOnBlock uses the item held in the main hand of the player on a block at the position passed. The
// player is assumed to have clicked the face passed with the relative click position clickPos.
// If the item could not be used successfully, for example when the position is out of range, the method
// returns immediately.
// UseItemOnBlock does nothing if the block at the cube.Pos passed is of the type block.Air.
func (p *Player) UseItemOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	if _, ok := p.tx.Block(pos).(block.Air); ok || !p.canReach(pos.Vec3Centre()) {
		// The client used its item on a block that does not exist server-side or one it couldn't reach. Stop trying
		// to use the item immediately.
		p.resendBlocks(pos, face)
		return
	}
	ctx := event.C(p)
	if p.Handler().HandleItemUseOnBlock(ctx, pos, face, clickPos); ctx.Cancelled() {
		p.resendBlocks(pos, face)
		return
	}
	i, left := p.HeldItems()
	b := p.tx.Block(pos)
	if act, ok := b.(block.Activatable); ok {
		// If a player is sneaking, it will not activate the block clicked, unless it is not holding any
		// items, in which case the block will be activated as usual.
		if !p.Sneaking() || i.Empty() {
			p.SwingArm()

			// The block was activated: Blocks such as doors must always have precedence over the item being
			// used.
			if useCtx := p.useContext(); act.Activate(pos, face, p.tx, p, useCtx) {
				p.SetHeldItems(p.subtractItem(p.damageItem(i, useCtx.Damage), useCtx.CountSub), left)
				p.addNewItem(useCtx)
				return
			}
		}
	}
	if i.Empty() {
		return
	}
	switch ib := i.Item().(type) {
	case item.UsableOnBlock:
		// The item does something when used on a block.
		useCtx := p.useContext()
		if !ib.UseOnBlock(pos, face, clickPos, p.tx, p, useCtx) {
			return
		}
		p.SwingArm()
		p.SetHeldItems(p.subtractItem(p.damageItem(i, useCtx.Damage), useCtx.CountSub), left)
		p.addNewItem(useCtx)
	case world.Block:
		// The item IS a block, meaning it is being placed.
		replacedPos := pos
		if replaceable, ok := b.(block.Replaceable); !ok || !replaceable.ReplaceableBy(ib) {
			// The block clicked was either not replaceable, or not replaceable using the block passed.
			replacedPos = pos.Side(face)
		}
		if replaceable, ok := p.tx.Block(replacedPos).(block.Replaceable); !ok || !replaceable.ReplaceableBy(ib) || replacedPos.OutOfBounds(p.tx.Range()) {
			return
		}
		if !p.placeBlock(replacedPos, ib, false) || p.GameMode().CreativeInventory() {
			return
		}
		p.SetHeldItems(p.subtractItem(i, 1), left)
	}
}

// UseItemOnEntity uses the item held in the main hand of the player on the entity passed, provided it is
// within range of the player.
// If the item held in the main hand of the player does nothing when used on an entity, nothing will happen.
func (p *Player) UseItemOnEntity(e world.Entity) bool {
	if !p.canReach(e.Position()) {
		return false
	}
	ctx := event.C(p)
	if p.Handler().HandleItemUseOnEntity(ctx, e); ctx.Cancelled() {
		return false
	}
	i, left := p.HeldItems()
	usable, ok := i.Item().(item.UsableOnEntity)
	if !ok {
		return true
	}
	useCtx := p.useContext()
	if !usable.UseOnEntity(e, p.tx, p, useCtx) {
		return true
	}
	p.SwingArm()
	p.SetHeldItems(p.subtractItem(p.damageItem(i, useCtx.Damage), useCtx.CountSub), left)
	p.addNewItem(useCtx)
	return true
}

// AttackEntity uses the item held in the main hand of the player to attack the entity passed, provided it is
// within range of the player.
// The damage dealt to the entity will depend on the item held by the player and any effects the player may
// have.
// If the player cannot reach the entity at its position, the method returns immediately.
func (p *Player) AttackEntity(e world.Entity) bool {
	if !p.canReach(e.Position()) {
		return false
	}
	var (
		force, height  = 0.45, 0.3608
		_, slowFalling = p.Effect(effect.SlowFalling)
		_, blind       = p.Effect(effect.Blindness)
		critical       = !p.Sprinting() && !p.Flying() && p.FallDistance() > 0 && !slowFalling && !blind
	)

	ctx := event.C(p)
	if p.Handler().HandleAttackEntity(ctx, e, &force, &height, &critical); ctx.Cancelled() {
		return false
	}
	p.SwingArm()

	i, _ := p.HeldItems()
	living, ok := e.(entity.Living)
	if !ok {
		return false
	}

	dmg := i.AttackDamage()
	if strength, ok := p.Effect(effect.Strength); ok {
		dmg += dmg * effect.Strength.Multiplier(strength.Level())
	}
	if weakness, ok := p.Effect(effect.Weakness); ok {
		dmg -= dmg * effect.Weakness.Multiplier(weakness.Level())
	}
	if s, ok := i.Enchantment(enchantment.Sharpness); ok {
		dmg += enchantment.Sharpness.Addend(s.Level())
	}
	if critical {
		dmg *= 1.5
	}

	n, vulnerable := living.Hurt(dmg, entity.AttackDamageSource{Attacker: p})
	i, left := p.HeldItems()

	p.tx.PlaySound(entity.EyePosition(e), sound.Attack{Damage: !mgl64.FloatEqual(n, 0)})
	if !vulnerable {
		return true
	}
	if critical {
		for _, v := range p.tx.Viewers(living.Position()) {
			v.ViewEntityAction(living, entity.CriticalHitAction{})
		}
	}

	p.Exhaust(0.1)

	if k, ok := i.Enchantment(enchantment.Knockback); ok {
		inc := enchantment.Knockback.Force(k.Level())
		force += inc
		height += inc
	}
	living.KnockBack(p.Position(), force, height)

	if f, ok := i.Enchantment(enchantment.FireAspect); ok {
		if flammable, ok := living.(entity.Flammable); ok {
			flammable.SetOnFire(enchantment.FireAspect.Duration(f.Level()))
		}
	}

	if durable, ok := i.Item().(item.Durable); ok {
		p.SetHeldItems(p.damageItem(i, durable.DurabilityInfo().AttackDurability), left)
	}
	return true
}

// StartBreaking makes the player start breaking the block at the position passed using the item currently
// held in its main hand.
// If no block is present at the position, or if the block is out of range, StartBreaking will return
// immediately and the block will not be broken. StartBreaking will stop the breaking of any block that the
// player might be breaking before this method is called.
func (p *Player) StartBreaking(pos cube.Pos, face cube.Face) {
	p.AbortBreaking()
	if _, air := p.tx.Block(pos).(block.Air); air || !p.canReach(pos.Vec3Centre()) {
		// The block was either out of range or air, so it can't be broken by the player.
		return
	}
	if _, ok := p.tx.Block(pos.Side(face)).(block.Fire); ok {
		ctx := event.C(p)
		if p.Handler().HandleFireExtinguish(ctx, pos); ctx.Cancelled() {
			// Resend the block because on client side that was extinguished
			p.resendBlocks(pos, face)
			return
		}

		p.tx.SetBlock(pos.Side(face), nil, nil)
		p.tx.PlaySound(pos.Vec3(), sound.FireExtinguish{})
		return
	}

	held, _ := p.HeldItems()
	if _, ok := held.Item().(item.Sword); ok && p.GameMode().CreativeInventory() {
		// Can't break blocks with a sword in creative mode.
		return
	}
	// Note: We intentionally store this regardless of whether the breaking proceeds, so that we
	// can resend the block to the client when it tries to break the block regardless.
	p.breakingPos = pos

	ctx := event.C(p)
	if p.Handler().HandleStartBreak(ctx, pos); ctx.Cancelled() {
		return
	}
	if punchable, ok := p.tx.Block(pos).(block.Punchable); ok {
		punchable.Punch(pos, face, p.tx, p)
	}

	p.breaking, p.breakingFace = true, face
	p.SwingArm()

	if p.GameMode().CreativeInventory() {
		return
	}
	p.lastBreakDuration = p.breakTime(pos)
	for _, viewer := range p.viewers() {
		viewer.ViewBlockAction(pos, block.StartCrackAction{BreakTime: p.lastBreakDuration})
	}
}

// breakTime returns the time needed to break a block at the position passed, taking into account the item
// held, if the player is on the ground/underwater and if the player has any effects.
func (p *Player) breakTime(pos cube.Pos) time.Duration {
	held, _ := p.HeldItems()
	breakTime := block.BreakDuration(p.tx.Block(pos), held)
	if !p.OnGround() {
		breakTime *= 5
	}
	if _, ok := p.Armour().Helmet().Enchantment(enchantment.AquaAffinity); p.insideOfWater() && !ok {
		breakTime *= 5
	}
	for _, e := range p.Effects() {
		lvl := e.Level()
		switch e.Type() {
		case effect.Haste:
			breakTime = time.Duration(float64(breakTime) * effect.Haste.Multiplier(lvl))
		case effect.MiningFatigue:
			breakTime = time.Duration(float64(breakTime) * effect.MiningFatigue.Multiplier(lvl))
		case effect.ConduitPower:
			breakTime = time.Duration(float64(breakTime) * effect.ConduitPower.Multiplier(lvl))
		}
	}
	return breakTime
}

// FinishBreaking makes the player finish breaking the block it is currently breaking, or returns immediately
// if the player isn't breaking anything.
// FinishBreaking will stop the animation and break the block.
func (p *Player) FinishBreaking() {
	if !p.breaking {
		p.resendBlock(p.breakingPos)
		return
	}
	p.AbortBreaking()
	p.BreakBlock(p.breakingPos)
}

// AbortBreaking makes the player stop breaking the block it is currently breaking, or returns immediately
// if the player isn't breaking anything.
// Unlike FinishBreaking, AbortBreaking does not stop the animation.
func (p *Player) AbortBreaking() {
	if !p.breaking {
		return
	}
	p.breaking, p.breakCounter = false, 0
	for _, viewer := range p.viewers() {
		viewer.ViewBlockAction(p.breakingPos, block.StopCrackAction{})
	}
}

// ContinueBreaking makes the player continue breaking the block it started breaking after a call to
// Player.StartBreaking().
// The face passed is used to display particles on the side of the block broken.
func (p *Player) ContinueBreaking(face cube.Face) {
	if !p.breaking {
		return
	}
	pos := p.breakingPos
	b := p.tx.Block(pos)
	p.tx.AddParticle(pos.Vec3(), particle.PunchBlock{Block: b, Face: face})

	if p.breakCounter++; p.breakCounter%5 == 0 {
		p.SwingArm()

		// We send this sound only every so often. Vanilla doesn't send it every tick while breaking
		// either. Every 5 ticks seems accurate.
		p.tx.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: b})
	}
	if breakTime := p.breakTime(pos); breakTime != p.lastBreakDuration {
		for _, viewer := range p.viewers() {
			viewer.ViewBlockAction(pos, block.ContinueCrackAction{BreakTime: breakTime})
		}
		p.lastBreakDuration = breakTime
	}
}

// PlaceBlock makes the player place the block passed at the position passed, granted it is within the range
// of the player.
// An item.UseContext may be passed to obtain information on if the block placement was successful. (SubCount will
// be incremented). Nil may also be passed for the context parameter.
func (p *Player) PlaceBlock(pos cube.Pos, b world.Block, ctx *item.UseContext) {
	ignoreBBox := ctx != nil && ctx.IgnoreBBox
	if !p.placeBlock(pos, b, ignoreBBox) {
		return
	}
	if ctx != nil {
		ctx.CountSub++
	}
}

// placeBlock makes the player place the block passed at the position passed, granted it is within the range
// of the player. A bool is returned indicating if a block was placed successfully.
func (p *Player) placeBlock(pos cube.Pos, b world.Block, ignoreBBox bool) bool {
	if !p.canReach(pos.Vec3Centre()) || !p.GameMode().AllowsEditing() {
		p.resendBlocks(pos, cube.Faces()...)
		return false
	}
	if obstructed, selfOnly := p.obstructedPos(pos, b); obstructed && !ignoreBBox {
		if !selfOnly {
			// Only resend blocks if there were other entities blocking the
			// placement than the player itself. Resending blocks placed inside
			// the player itself leads to synchronisation issues.
			p.resendBlocks(pos, cube.Faces()...)
		}
		return false
	}

	ctx := event.C(p)
	if p.Handler().HandleBlockPlace(ctx, pos, b); ctx.Cancelled() {
		p.resendBlocks(pos, cube.Faces()...)
		return false
	}
	p.tx.SetBlock(pos, b, nil)
	p.tx.PlaySound(pos.Vec3(), sound.BlockPlace{Block: b})
	p.SwingArm()
	return true
}

// obstructedPos checks if the position passed is obstructed if the block
// passed is attempted to be placed. The function returns true as the first
// bool if there is an entity in the way that could prevent the block from
// being placed.
// If the only entity preventing the block from being placed is the player
// itself, the second bool returned is true too.
func (p *Player) obstructedPos(pos cube.Pos, b world.Block) (obstructed, selfOnly bool) {
	blockBoxes := b.Model().BBox(pos, p.tx)
	for i, box := range blockBoxes {
		blockBoxes[i] = box.Translate(pos.Vec3())
	}

	for e := range p.tx.EntitiesWithin(cube.Box(-3, -3, -3, 3, 3, 3).Translate(pos.Vec3())) {
		t := e.H().Type()
		switch t {
		case entity.ItemType, entity.ArrowType:
			continue
		default:
			if cube.AnyIntersections(blockBoxes, t.BBox(e).Translate(e.Position()).Grow(-1e-4)) {
				obstructed = true
				if e.H() == p.handle {
					continue
				}
				return true, false
			}
		}
	}
	return obstructed, true
}

// BreakBlock makes the player break a block in the world at a position passed. If the player is unable to
// reach the block passed, the method returns immediately.
func (p *Player) BreakBlock(pos cube.Pos) {
	b := p.tx.Block(pos)
	if _, air := b.(block.Air); air {
		// Don't do anything if the position broken is already air.
		return
	}
	if !p.canReach(pos.Vec3Centre()) || !p.GameMode().AllowsEditing() {
		p.resendBlocks(pos)
		return
	}
	if _, breakable := b.(block.Breakable); !breakable && !p.GameMode().CreativeInventory() {
		p.resendBlocks(pos)
		return
	}
	held, _ := p.HeldItems()
	drops := p.drops(held, b)

	xp := 0
	if breakable, ok := b.(block.Breakable); ok && !p.GameMode().CreativeInventory() {
		xp = breakable.BreakInfo().XPDrops.RandomValue()
	}

	ctx := event.C(p)
	if p.Handler().HandleBlockBreak(ctx, pos, &drops, &xp); ctx.Cancelled() {
		p.resendBlocks(pos)
		return
	}
	held, left := p.HeldItems()

	p.SwingArm()
	p.tx.SetBlock(pos, nil, nil)
	p.tx.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})

	if breakable, ok := b.(block.Breakable); ok {
		info := breakable.BreakInfo()
		if info.BreakHandler != nil {
			info.BreakHandler(pos, p.tx, p)
		}
		for _, orb := range entity.NewExperienceOrbs(pos.Vec3Centre(), xp) {
			p.tx.AddEntity(orb)
		}
	}
	for _, drop := range drops {
		opts := world.EntitySpawnOpts{Position: pos.Vec3Centre(), Velocity: mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1}}
		p.tx.AddEntity(entity.NewItem(opts, drop))
	}

	p.Exhaust(0.005)
	if block.BreaksInstantly(b, held) {
		return
	}
	if durable, ok := held.Item().(item.Durable); ok {
		p.SetHeldItems(p.damageItem(held, durable.DurabilityInfo().BreakDurability), left)
	}
}

// drops returns the drops that the player can get from the block passed using the item held.
func (p *Player) drops(held item.Stack, b world.Block) []item.Stack {
	t, ok := held.Item().(item.Tool)
	if !ok {
		t = item.ToolNone{}
	}
	var drops []item.Stack
	if breakable, ok := b.(block.Breakable); ok && !p.GameMode().CreativeInventory() {
		if breakable.BreakInfo().Harvestable(t) {
			drops = breakable.BreakInfo().Drops(t, held.Enchantments())
		}
	} else if it, ok := b.(world.Item); ok && !p.GameMode().CreativeInventory() {
		drops = []item.Stack{item.NewStack(it, 1)}
	}
	return drops
}

// PickBlock makes the player pick a block in the world at a position passed. If the player is unable to
// pick the block, the method returns immediately.
func (p *Player) PickBlock(pos cube.Pos) {
	if !p.canReach(pos.Vec3()) {
		return
	}

	b := p.tx.Block(pos)

	var pickedItem item.Stack
	if pi, ok := b.(block.Pickable); ok {
		pickedItem = pi.Pick()
	} else if i, ok := b.(world.Item); ok {
		it, _ := world.ItemByName(i.EncodeItem())
		pickedItem = item.NewStack(it, 1)
	} else {
		return
	}

	slot, found := p.Inventory().First(pickedItem)
	if !found && !p.GameMode().CreativeInventory() {
		return
	}

	ctx := event.C(p)
	if p.Handler().HandleBlockPick(ctx, pos, b); ctx.Cancelled() {
		return
	}
	_, offhand := p.HeldItems()

	if found {
		if slot < 9 {
			_ = p.SetHeldSlot(slot)
			return
		}
		_ = p.Inventory().Swap(slot, int(*p.heldSlot))
		return
	}

	firstEmpty, emptyFound := p.Inventory().FirstEmpty()
	if !emptyFound {
		p.SetHeldItems(pickedItem, offhand)
		return
	}
	if firstEmpty < 9 {
		_ = p.SetHeldSlot(firstEmpty)
		_ = p.Inventory().SetItem(firstEmpty, pickedItem)
		return
	}
	_ = p.Inventory().Swap(firstEmpty, int(*p.heldSlot))
	p.SetHeldItems(pickedItem, offhand)
}

// Teleport teleports the player to a target position in the world. Unlike Move, it immediately changes the
// position of the player, rather than showing an animation.
func (p *Player) Teleport(pos mgl64.Vec3) {
	ctx := event.C(p)
	if p.Handler().HandleTeleport(ctx, pos); ctx.Cancelled() {
		return
	}
	p.teleport(pos)
}

// teleport teleports the player to a target position in the world. It does not call the Handler of the
// player.
func (p *Player) teleport(pos mgl64.Vec3) {
	for _, v := range p.viewers() {
		v.ViewEntityTeleport(p, pos)
	}
	p.data.Pos = pos
	p.data.Vel = mgl64.Vec3{}
	p.ResetFallDistance()
}

// Move moves the player from one position to another in the world, by adding the delta passed to the current
// position of the player.
// Move also rotates the player, adding deltaYaw and deltaPitch to the respective values.
func (p *Player) Move(deltaPos mgl64.Vec3, deltaYaw, deltaPitch float64) {
	if p.Dead() || (deltaPos.ApproxEqual(mgl64.Vec3{}) && mgl64.FloatEqual(deltaYaw, 0) && mgl64.FloatEqual(deltaPitch, 0)) {
		return
	}
	if p.immobile {
		if mgl64.FloatEqual(deltaYaw, 0) && mgl64.FloatEqual(deltaPitch, 0) {
			// If only the position was changed, don't continue with the movement when immobile.
			return
		}
		// Still update rotation if it was changed.
		deltaPos = mgl64.Vec3{}
	}
	var (
		pos         = p.Position()
		res, resRot = pos.Add(deltaPos), p.Rotation().Add(cube.Rotation{deltaYaw, deltaPitch})
	)
	ctx := event.C(p)
	if p.Handler().HandleMove(ctx, res, resRot); ctx.Cancelled() {
		if p.session() != session.Nop && pos.ApproxEqual(p.Position()) {
			// The position of the player was changed and the event cancelled. This means we still need to notify the
			// player of this movement change.
			p.teleport(pos)
		}
		return
	}
	for _, v := range p.viewers() {
		v.ViewEntityMovement(p, res, resRot, p.OnGround())
	}

	p.data.Pos = res
	p.data.Rot = resRot
	if deltaPos.Len() <= 3 {
		// Only update velocity if the player is not moving too fast to prevent potential OOMs.
		p.data.Vel = deltaPos
		p.checkBlockCollisions(deltaPos)
	}

	horizontalVel := deltaPos
	horizontalVel[1] = 0
	if p.Gliding() {
		if deltaPos.Y() >= -0.5 {
			p.fallDistance = 1.0
		}
		if p.collidedHorizontally {
			if force := horizontalVel.Len()*10.0 - 3.0; force > 0.0 {
				p.tx.PlaySound(p.Position(), sound.Fall{Distance: force})
				p.Hurt(force, entity.GlideDamageSource{})
			}
		}
	}

	_, submergedBefore := p.tx.Liquid(cube.PosFromVec3(pos.Add(mgl64.Vec3{0, p.EyeHeight()})))
	_, submergedAfter := p.tx.Liquid(cube.PosFromVec3(res.Add(mgl64.Vec3{0, p.EyeHeight()})))
	if submergedBefore != submergedAfter {
		// Player wasn't either breathing before and no longer isn't, or wasn't breathing before and now is,
		// so send the updated metadata.
		p.session().ViewEntityState(p)
	}

	p.onGround = p.checkOnGround(deltaPos)
	p.updateFallState(deltaPos[1])

	if p.Swimming() {
		p.Exhaust(0.01 * horizontalVel.Len())
	} else if p.Sprinting() {
		p.Exhaust(0.1 * horizontalVel.Len())
	}
}

// Position returns the current position of the player. It may be changed as the player moves or is moved
// around the world.
func (p *Player) Position() mgl64.Vec3 {
	return p.data.Pos
}

// Velocity returns the players current velocity. If there is an attached session, this will be empty.
func (p *Player) Velocity() mgl64.Vec3 {
	return p.data.Vel
}

// SetVelocity updates the player's velocity. If there is an attached session, this will just send
// the velocity to the player session for the player to update.
func (p *Player) SetVelocity(velocity mgl64.Vec3) {
	if p.session() == session.Nop {
		p.data.Vel = velocity
		return
	}
	for _, v := range p.viewers() {
		v.ViewEntityVelocity(p, velocity)
	}
}

// Rotation returns the yaw and pitch of the player in degrees. Yaw is horizontal rotation (rotation around the
// vertical axis, 0 when facing forward), pitch is vertical rotation (rotation around the horizontal axis, also 0
// when facing forward).
func (p *Player) Rotation() cube.Rotation {
	return p.data.Rot
}

// Collect makes the player collect the item stack passed, adding it to the inventory. The amount of items that could
// be added is returned.
func (p *Player) Collect(s item.Stack) (int, bool) {
	if p.Dead() || !p.GameMode().AllowsInteraction() {
		return 0, false
	}
	ctx := event.C(p)
	if p.Handler().HandleItemPickup(ctx, &s); ctx.Cancelled() {
		return 0, false
	}
	var added int
	if _, offHand := p.HeldItems(); !offHand.Empty() && offHand.Comparable(s) {
		added, _ = p.offHand.AddItem(s)
	}
	if s.Count() != added {
		n, _ := p.Inventory().AddItem(s.Grow(-added))
		added += n
	}
	return added, true
}

// Experience returns the amount of experience the player has.
func (p *Player) Experience() int {
	return p.experience.Experience()
}

// EnchantmentSeed is a seed used to calculate random enchantments with enchantment tables.
func (p *Player) EnchantmentSeed() int64 {
	return p.enchantSeed
}

// ResetEnchantmentSeed resets the enchantment seed to a new random value.
func (p *Player) ResetEnchantmentSeed() {
	p.enchantSeed = rand.Int64()
}

// AddExperience adds experience to the player.
func (p *Player) AddExperience(amount int) int {
	ctx := event.C(p)
	if p.Handler().HandleExperienceGain(ctx, &amount); ctx.Cancelled() {
		return 0
	}
	before := p.experience.Level()
	level, _ := p.experience.Add(amount)
	if level/5 > before/5 {
		p.PlaySound(sound.LevelUp{})
	} else if amount > 0 {
		p.PlaySound(sound.Experience{})
	}
	p.session().SendExperience(p.ExperienceLevel(), p.ExperienceProgress())
	return amount
}

// RemoveExperience removes experience from the player.
func (p *Player) RemoveExperience(amount int) {
	p.experience.Add(-amount)
	p.session().SendExperience(p.ExperienceLevel(), p.ExperienceProgress())
}

// ExperienceLevel returns the experience level of the player.
func (p *Player) ExperienceLevel() int {
	return p.experience.Level()
}

// SetExperienceLevel sets the experience level of the player. The level must have a value between 0 and 2,147,483,647,
// otherwise the method panics.
func (p *Player) SetExperienceLevel(level int) {
	p.experience.SetLevel(level)
	p.session().SendExperience(p.ExperienceLevel(), p.ExperienceProgress())
}

// ExperienceProgress returns the experience progress of the player.
func (p *Player) ExperienceProgress() float64 {
	return p.experience.Progress()
}

// SetExperienceProgress sets the experience progress of the player. The progress must have a value between 0.0 and 1.0, otherwise
// the method panics.
func (p *Player) SetExperienceProgress(progress float64) {
	p.experience.SetProgress(progress)
	p.session().SendExperience(p.ExperienceLevel(), p.ExperienceProgress())
}

// CollectExperience makes the player collect the experience points passed, adding it to the experience manager. A bool
// is returned indicating whether the player was able to collect the experience or not, due to the 100ms delay between
// experience collection or if the player was dead or in a game mode that doesn't allow collection.
func (p *Player) CollectExperience(value int) bool {
	if p.Dead() || !p.GameMode().AllowsInteraction() {
		return false
	}
	if last := p.lastXPPickup; last != nil && time.Since(*last) < time.Millisecond*100 {
		return false
	}
	value = p.mendItems(value)
	now := time.Now()
	p.lastXPPickup = &now
	if value > 0 {
		return p.AddExperience(value) > 0
	}

	p.PlaySound(sound.Experience{})
	return true
}

// mendItems handles the mending enchantment when collecting experience, it then returns the leftover experience.
func (p *Player) mendItems(xp int) int {
	mendingItems := make([]item.Stack, 0, 6)
	held, offHand := p.HeldItems()
	if _, ok := offHand.Enchantment(enchantment.Mending); ok && offHand.Durability() < offHand.MaxDurability() {
		mendingItems = append(mendingItems, offHand)
	}
	if _, ok := held.Enchantment(enchantment.Mending); ok && held.Durability() < held.MaxDurability() {
		mendingItems = append(mendingItems, held)
	}
	for _, i := range p.Armour().Items() {
		if i.Durability() == i.MaxDurability() {
			continue
		}
		if _, ok := i.Enchantment(enchantment.Mending); ok {
			mendingItems = append(mendingItems, i)
		}
	}
	length := len(mendingItems)
	if length == 0 {
		return xp
	}
	foundItem := mendingItems[rand.IntN(length)]
	repairAmount := math.Min(float64(foundItem.MaxDurability()-foundItem.Durability()), float64(xp*2))
	repairedItem := foundItem.WithDurability(foundItem.Durability() + int(repairAmount))
	if repairAmount >= 2 {
		// mending removes 1 experience point for every 2 durability points. If the repaired durability is less than 2,
		// then no experience is removed.
		xp -= int(math.Ceil(repairAmount / 2))
	}
	if offHand.Equal(foundItem) {
		p.SetHeldItems(held, repairedItem)
	} else if held.Equal(foundItem) {
		p.SetHeldItems(repairedItem, offHand)
	} else if slot, ok := p.Armour().Inventory().First(foundItem); ok {
		_ = p.Armour().Inventory().SetItem(slot, repairedItem)
	}
	return xp
}

// Drop makes the player drop the item.Stack passed as an entity.Item, so that it may be picked up from the
// ground.
// The dropped item entity has a pickup delay of 2 seconds.
// The number of items that was dropped in the end is returned. It is generally the count of the stack passed
// or 0 if dropping the item.Stack was cancelled.
func (p *Player) Drop(s item.Stack) int {
	ctx := event.C(p)
	if p.Handler().HandleItemDrop(ctx, s); ctx.Cancelled() {
		return 0
	}
	opts := world.EntitySpawnOpts{Position: p.Position().Add(mgl64.Vec3{0, 1.4}), Velocity: p.Rotation().Vec3().Mul(0.4)}
	p.tx.AddEntity(entity.NewItemPickupDelay(opts, s, time.Second*2))
	return s.Count()
}

// OpenBlockContainer opens a block container, such as a chest, at the position passed. If no container was
// present at that location, OpenBlockContainer does nothing.
// OpenBlockContainer will also do nothing if the player has no session connected to it.
func (p *Player) OpenBlockContainer(pos cube.Pos, tx *world.Tx) {
	if p.session() != session.Nop {
		p.session().OpenBlockContainer(pos, tx)
	}
}

// HideEntity hides a world.Entity from the Player so that it can under no circumstance see it. Hidden entities can be
// made visible again through a call to ShowEntity.
func (p *Player) HideEntity(e world.Entity) {
	if p.session() != session.Nop && p.H() != e.H() {
		p.session().StopShowingEntity(e)
	}
}

// ShowEntity shows a world.Entity previously hidden from the Player using HideEntity. It does nothing if the entity
// wasn't currently hidden.
func (p *Player) ShowEntity(e world.Entity) {
	if p.session() != session.Nop {
		p.session().StartShowingEntity(e)
	}
}

// Latency returns a rolling average of latency between the sending and the receiving end of the connection of
// the player.
// The latency returned is updated continuously and is half the round trip time (RTT).
// If the Player does not have a session associated with it, Latency returns 0.
func (p *Player) Latency() time.Duration {
	if p.session() == session.Nop {
		return 0
	}
	return p.session().Latency()
}

// Tick ticks the entity, performing actions such as checking if the player is still breaking a block.
func (p *Player) Tick(tx *world.Tx, current int64) {
	if p.Dead() {
		return
	}
	if _, ok := p.tx.Liquid(cube.PosFromVec3(p.Position())); !ok {
		p.StopSwimming()
		if _, ok := p.Armour().Helmet().Item().(item.TurtleShell); ok {
			p.AddEffect(effect.New(effect.WaterBreathing, 1, time.Second*10).WithoutParticles())
		}
	}

	if _, ok := p.Armour().Chestplate().Item().(item.Elytra); ok && p.Gliding() {
		if p.glideTicks += 1; p.glideTicks%20 == 0 {
			d := p.damageItem(p.Armour().Chestplate(), 1)
			p.armour.SetChestplate(d)
			if d.Durability() < 2 {
				p.StopGliding()
			}
		}
	}

	p.checkBlockCollisions(p.data.Vel)
	p.onGround = p.checkOnGround(mgl64.Vec3{})

	p.effects.Tick(p, p.tx)

	p.tickFood()
	p.tickAirSupply()

	if p.Position()[1] < float64(p.tx.Range()[0]) {
		p.Hurt(4, entity.VoidDamageSource{})
	}
	if p.insideOfSolid() {
		p.Hurt(1, entity.SuffocationDamageSource{})
	}

	if p.OnFireDuration() > 0 {
		p.fireTicks -= 1
		if !p.GameMode().AllowsTakingDamage() || p.OnFireDuration() <= 0 || p.tx.RainingAt(cube.PosFromVec3(p.Position())) {
			p.Extinguish()
		}
		if p.OnFireDuration()%time.Second == 0 {
			p.Hurt(1, block.FireDamageSource{})
		}
	}

	held, _ := p.HeldItems()
	if current%4 == 0 && p.usingItem {
		if _, ok := held.Item().(item.Consumable); ok {
			// Eating particles seem to happen roughly every 4 ticks.
			for _, v := range p.viewers() {
				v.ViewEntityAction(p, entity.EatAction{})
			}
		}
	}

	if p.usingItem {
		if c, ok := held.Item().(item.Chargeable); ok {
			c.ContinueCharge(p, tx, p.useContext(), p.useDuration())
		}
	}
	if p.breaking {
		p.ContinueBreaking(p.breakingFace)
	}

	for it, ti := range p.cooldowns {
		if time.Now().After(ti) {
			delete(p.cooldowns, it)
		}
	}

	p.s.SendDebugShapes()

	if p.prevWorld != tx.World() && p.prevWorld != nil {
		p.Handler().HandleChangeWorld(p, p.prevWorld, tx.World())
	}
	p.prevWorld = tx.World()

	if p.session() == session.Nop && !p.Immobile() {
		m := p.mc.TickMovement(p, p.Position(), p.Velocity(), p.Rotation(), p.tx)
		m.Send()

		p.data.Vel = m.Velocity()
		p.Move(m.Position().Sub(p.Position()), 0, 0)
	} else {
		p.data.Vel = mgl64.Vec3{}
	}
}

// tickAirSupply tick's the player's air supply, consuming it when underwater, and replenishing it when out of water.
func (p *Player) tickAirSupply() {
	if !p.canBreathe() {
		if r, ok := p.Armour().Helmet().Enchantment(enchantment.Respiration); ok && rand.Float64() <= enchantment.Respiration.Chance(r.Level()) {
			// respiration grants a chance to avoid drowning damage every tick.
			return
		}
		if p.airSupplyTicks -= 1; p.airSupplyTicks <= -20 {
			p.airSupplyTicks = 0
			p.Hurt(2, entity.DrowningDamageSource{})
		}
		p.breathing = false
		p.updateState()
	} else if !p.breathing && p.airSupplyTicks < p.maxAirSupplyTicks {
		p.airSupplyTicks = min(p.airSupplyTicks+5, p.maxAirSupplyTicks)
		p.breathing = p.airSupplyTicks == p.maxAirSupplyTicks
		p.updateState()
	}
}

// tickFood ticks food related functionality, such as the depletion of the food bar and regeneration if it
// is full enough.
func (p *Player) tickFood() {
	if p.hunger.foodTick%10 == 0 && (p.hunger.canQuicklyRegenerate() || p.tx.World().Difficulty().FoodRegenerates()) {
		if p.tx.World().Difficulty().FoodRegenerates() {
			p.AddFood(1)
		}
		if p.hunger.foodTick%20 == 0 {
			p.regenerate(true)
		}
	}
	if p.hunger.foodTick == 1 {
		if p.hunger.canRegenerate() {
			p.regenerate(false)
		} else if p.hunger.starving() {
			p.starve()
		}
	}

	if !p.hunger.canSprint() {
		p.StopSprinting()
	}

	p.hunger.foodTick++
	if p.hunger.foodTick > 80 {
		p.hunger.foodTick = 1
	}
}

// regenerate attempts to regenerate half a heart of health, typically caused by a full food bar.
func (p *Player) regenerate(exhaust bool) {
	if p.Health() == p.MaxHealth() {
		return
	}
	p.Heal(1, entity.FoodHealingSource{})
	if exhaust {
		p.Exhaust(6)
	}
}

// starve deals starvation damage to the player if the difficult allows it. In peaceful mode, no damage will
// ever be dealt. In easy mode, damage will only be dealt if the player has more than 10 health. In normal
// mode, damage will only be dealt if the player has more than 2 health and in hard mode, damage will always
// be dealt.
func (p *Player) starve() {
	if p.Health() > p.tx.World().Difficulty().StarvationHealthLimit() {
		p.Hurt(1, StarvationDamageSource{})
	}
}

// AirSupply returns the player's remaining air supply.
func (p *Player) AirSupply() time.Duration {
	return time.Duration(p.airSupplyTicks) * time.Second / 20
}

// SetAirSupply sets the player's remaining air supply.
func (p *Player) SetAirSupply(duration time.Duration) {
	p.airSupplyTicks = int(duration.Milliseconds() / 50)
	p.updateState()
}

// MaxAirSupply returns the player's maximum air supply.
func (p *Player) MaxAirSupply() time.Duration {
	return time.Duration(p.maxAirSupplyTicks) * time.Second / 20
}

// SetMaxAirSupply sets the player's maximum air supply.
func (p *Player) SetMaxAirSupply(duration time.Duration) {
	p.maxAirSupplyTicks = int(duration.Milliseconds() / 50)
	p.updateState()
}

// canBreathe returns true if the player can currently breathe.
func (p *Player) canBreathe() bool {
	canTakeDamage := p.GameMode().AllowsTakingDamage()
	_, waterBreathing := p.effects.Effect(effect.WaterBreathing)
	_, conduitPower := p.effects.Effect(effect.ConduitPower)
	return !canTakeDamage || waterBreathing || conduitPower || (!p.insideOfWater() && !p.insideOfSolid())
}

// breathingDistanceBelowEyes is the lowest distance the player can be in water and still be able to breathe based on
// the player's eye height.
const breathingDistanceBelowEyes = 0.11111111

// insideOfWater returns true if the player is currently underwater.
func (p *Player) insideOfWater() bool {
	pos := cube.PosFromVec3(entity.EyePosition(p))
	if l, ok := p.tx.Liquid(pos); ok {
		if _, ok := l.(block.Water); ok {
			d := float64(l.SpreadDecay()) + 1
			if l.LiquidFalling() {
				d = 1
			}
			return p.Position().Y() < (pos.Side(cube.FaceUp).Vec3().Y())-(d/9-breathingDistanceBelowEyes)
		}
	}
	return false
}

// insideOfSolid returns true if the player is inside a solid block.
func (p *Player) insideOfSolid() bool {
	pos := cube.PosFromVec3(entity.EyePosition(p))
	b, box := p.tx.Block(pos), p.handle.Type().BBox(p).Translate(p.Position())

	_, solid := b.Model().(model.Solid)
	if !solid {
		// Not solid.
		return false
	}
	d, diffuses := b.(block.LightDiffuser)
	if diffuses && d.LightDiffusionLevel() == 0 {
		// Transparent.
		return false
	}
	for _, blockBox := range b.Model().BBox(pos, p.tx) {
		if blockBox.Translate(pos.Vec3()).IntersectsWith(box) {
			return true
		}
	}
	return false
}

// checkCollisions checks the player's block collisions.
func (p *Player) checkBlockCollisions(vel mgl64.Vec3) {
	entityBBox := Type.BBox(p).Translate(p.Position())
	deltaX, deltaY, deltaZ := vel[0], vel[1], vel[2]

	p.checkEntityInsiders(entityBBox)

	grown := entityBBox.Extend(vel).Grow(0.25)
	low, high := grown.Min(), grown.Max()
	minX, minY, minZ := int(math.Floor(low[0])), int(math.Floor(low[1])), int(math.Floor(low[2]))
	maxX, maxY, maxZ := int(math.Ceil(high[0])), int(math.Ceil(high[1])), int(math.Ceil(high[2]))

	// A prediction of one BBox per block, plus an additional 2, in case
	blocks := make([]cube.BBox, 0, (maxX-minX)*(maxY-minY)*(maxZ-minZ)+2)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				pos := cube.Pos{x, y, z}
				boxes := p.tx.Block(pos).Model().BBox(pos, p.tx)
				for _, box := range boxes {
					blocks = append(blocks, box.Translate(pos.Vec3()))
				}
			}
		}
	}

	// epsilon is the epsilon used for thresholds for change used for change in position and velocity.
	const epsilon = 0.001

	if !mgl64.FloatEqualThreshold(deltaY, 0, epsilon) {
		// First we move the entity BBox on the Y axis.
		for _, blockBBox := range blocks {
			deltaY = entityBBox.YOffset(blockBBox, deltaY)
		}
		entityBBox = entityBBox.Translate(mgl64.Vec3{0, deltaY})
	}
	if !mgl64.FloatEqualThreshold(deltaX, 0, epsilon) {
		// Then on the X axis.
		for _, blockBBox := range blocks {
			deltaX = entityBBox.XOffset(blockBBox, deltaX)
		}
		entityBBox = entityBBox.Translate(mgl64.Vec3{deltaX})
	}
	if !mgl64.FloatEqualThreshold(deltaZ, 0, epsilon) {
		// And finally on the Z axis.
		for _, blockBBox := range blocks {
			deltaZ = entityBBox.ZOffset(blockBBox, deltaZ)
		}
	}

	p.collidedHorizontally = !mgl64.FloatEqual(deltaX, vel[0]) || !mgl64.FloatEqual(deltaZ, vel[2])
	p.collidedVertically = !mgl64.FloatEqual(deltaY, vel[1])
}

// checkEntityInsiders checks if the player is colliding with any EntityInsider blocks.
func (p *Player) checkEntityInsiders(entityBBox cube.BBox) {
	box := entityBBox.Grow(-0.0001)
	low, high := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())

	for y := low[1]; y <= high[1]; y++ {
		for x := low[0]; x <= high[0]; x++ {
			for z := low[2]; z <= high[2]; z++ {
				blockPos := cube.Pos{x, y, z}
				b := p.tx.Block(blockPos)
				if collide, ok := b.(block.EntityInsider); ok {
					collide.EntityInside(blockPos, p.tx, p)
					if _, liquid := b.(world.Liquid); liquid {
						continue
					}
				}

				if l, ok := p.tx.Liquid(blockPos); ok {
					if collide, ok := l.(block.EntityInsider); ok {
						collide.EntityInside(blockPos, p.tx, p)
					}
				}
			}
		}
	}
}

// checkOnGround checks if the player is currently considered to be on the ground.
func (p *Player) checkOnGround(deltaPos mgl64.Vec3) bool {
	box := Type.BBox(p).Translate(p.Position()).Extend(mgl64.Vec3{0, -0.05}).Extend(deltaPos.Mul(-1.0))
	b := box.Grow(1)

	epsilon := mgl64.Vec3{mgl64.Epsilon, mgl64.Epsilon, mgl64.Epsilon}
	low, high := cube.PosFromVec3(b.Min().Add(epsilon)), cube.PosFromVec3(b.Max().Sub(epsilon))
	for x := low[0]; x <= high[0]; x++ {
		for z := low[2]; z <= high[2]; z++ {
			for y := low[1]; y < high[1]; y++ {
				pos := cube.Pos{x, y, z}
				for _, bb := range p.tx.Block(pos).Model().BBox(pos, p.tx) {
					if bb.Translate(pos.Vec3()).IntersectsWith(box) {
						return true
					}
				}
			}
		}
	}
	return false
}

// Scale returns the scale modifier of the Player. The default value for a normal scale is 1. A scale of 0
// will make the Player completely invisible.
func (p *Player) Scale() float64 {
	return p.scale
}

// SetScale changes the scale modifier of the Player. The default value for a normal scale is 1. A scale of 0
// will make the Player completely invisible.
func (p *Player) SetScale(s float64) {
	p.scale = s
	p.updateState()
}

// OnGround checks if the player is considered to be on the ground.
func (p *Player) OnGround() bool {
	if p.session() == session.Nop {
		return p.mc.OnGround()
	}
	return p.onGround
}

// EyeHeight returns the eye height of the player: 1.62, 1.26 if player is sneaking or 0.52 if the player is
// swimming, gliding or crawling.
func (p *Player) EyeHeight() float64 {
	switch {
	case p.swimming || p.crawling || p.gliding:
		return 0.52
	case p.sneaking:
		return 1.26
	default:
		return 1.62
	}
}

// TorsoHeight returns the torso height of the player: 1.52, 1.16 if the player is sneaking, or 0.42 if the player is
// swimming, gliding, or crawling.
func (p *Player) TorsoHeight() float64 {
	switch {
	case p.swimming || p.crawling || p.gliding:
		return 0.42
	case p.sneaking:
		return 1.16
	default:
		return 1.52
	}
}

// PlaySound plays a world.Sound that only this Player can hear. Unlike World.PlaySound, it is not broadcast
// to players around it.
func (p *Player) PlaySound(sound world.Sound) {
	p.session().PlaySound(sound, entity.EyePosition(p))
}

// ShowParticle shows a particle that only this Player can see. Unlike World.AddParticle, it is not broadcast
// to players around it.
func (p *Player) ShowParticle(pos mgl64.Vec3, particle world.Particle) {
	p.session().ViewParticle(pos, particle)
}

// OpenSign makes the player open the sign at the cube.Pos passed, with the specific side provided. The client will not
// show the interface if it is not aware of a sign at the position.
func (p *Player) OpenSign(pos cube.Pos, frontSide bool) {
	p.session().OpenSign(pos, frontSide)
}

// EditSign edits the sign at the cube.Pos passed and writes the text passed to a sign at that position. If no sign is
// present, an error is returned.
func (p *Player) EditSign(pos cube.Pos, frontText, backText string) error {
	sign, ok := p.tx.Block(pos).(block.Sign)
	if !ok {
		return fmt.Errorf("edit sign: no sign at position %v", pos)
	}

	if sign.Waxed {
		return nil
	} else if frontText == sign.Front.Text && backText == sign.Back.Text {
		return nil
	}

	ctx := event.C(p)
	if frontText != sign.Front.Text {
		if p.Handler().HandleSignEdit(ctx, pos, true, sign.Front.Text, frontText); ctx.Cancelled() {
			p.resendBlock(pos)
			return nil
		}
		sign.Front.Text = frontText
		sign.Front.Owner = p.XUID()
	} else {
		if p.Handler().HandleSignEdit(ctx, pos, false, sign.Back.Text, backText); ctx.Cancelled() {
			p.resendBlock(pos)
			return nil
		}
		sign.Back.Text = backText
		sign.Back.Owner = p.XUID()
	}
	p.tx.SetBlock(pos, sign, nil)
	return nil
}

// TurnLecternPage edits the lectern at the cube.Pos passed by turning the page to the page passed. If no lectern is
// present, an error is returned.
func (p *Player) TurnLecternPage(pos cube.Pos, page int) error {
	lectern, ok := p.tx.Block(pos).(block.Lectern)
	if !ok {
		return fmt.Errorf("edit lectern: no lectern at position %v", pos)
	}

	ctx := event.C(p)
	if p.Handler().HandleLecternPageTurn(ctx, pos, lectern.Page, &page); ctx.Cancelled() {
		return nil
	}

	lectern.Page = page
	p.tx.SetBlock(pos, lectern, nil)
	return nil
}

// updateState updates the state of the player to all viewers of the player.
func (p *Player) updateState() {
	for _, v := range p.viewers() {
		v.ViewEntityState(p)
	}
}

// Breathing checks if the player is currently able to breathe. If it's underwater and the player does not
// have the water breathing or conduit power effect, this returns false.
// If the player is in creative or spectator mode, Breathing always returns true.
func (p *Player) Breathing() bool {
	_, breathing := p.Effect(effect.WaterBreathing)
	_, conduitPower := p.Effect(effect.ConduitPower)
	_, submerged := p.tx.Liquid(cube.PosFromVec3(entity.EyePosition(p)))
	return !p.GameMode().AllowsTakingDamage() || !submerged || breathing || conduitPower
}

// SwingArm makes the player swing its arm.
func (p *Player) SwingArm() {
	if p.Dead() {
		return
	}
	for _, v := range p.viewers() {
		v.ViewEntityAction(p, entity.SwingArmAction{})
	}
}

// PunchAir makes the player punch the air and plays the sound for attacking with no damage.
func (p *Player) PunchAir() {
	if p.Dead() {
		return
	}
	ctx := event.C(p)
	if p.Handler().HandlePunchAir(ctx); ctx.Cancelled() {
		return
	}
	p.SwingArm()
	p.tx.PlaySound(p.Position(), sound.Attack{})
}

// UpdateDiagnostics updates the diagnostics of the player.
func (p *Player) UpdateDiagnostics(d session.Diagnostics) {
	p.Handler().HandleDiagnostics(p, d)
}

// AddDebugShape adds a debug shape to be rendered to the player. If the shape already exists, it will be
// updated with the new information.
func (p *Player) AddDebugShape(shape debug.Shape) {
	p.s.AddDebugShape(shape)
}

// RemoveDebugShape removes a debug shape from the player by its unique identifier.
func (p *Player) RemoveDebugShape(shape debug.Shape) {
	p.s.RemoveDebugShape(shape)
}

// VisibleDebugShapes returns a slice of all debug shapes that are currently being shown to the player.
func (p *Player) VisibleDebugShapes() []debug.Shape {
	return p.s.VisibleDebugShapes()
}

// RemoveAllDebugShapes removes all rendered debug shapes from the player, as well as any shapes that have
// not yet been rendered.
func (p *Player) RemoveAllDebugShapes() {
	p.s.RemoveAllDebugShapes()
}

// damageItem damages the item stack passed with the damage passed and returns the new stack. If the item
// broke, a breaking sound is played.
// If the player is not survival, the original stack is returned.
func (p *Player) damageItem(s item.Stack, d int) item.Stack {
	if p.GameMode().CreativeInventory() || d == 0 || s.MaxDurability() == -1 {
		return s
	}
	ctx := event.C(p)
	if p.Handler().HandleItemDamage(ctx, s, d); ctx.Cancelled() {
		return s
	}
	if e, ok := s.Enchantment(enchantment.Unbreaking); ok {
		d = enchantment.Unbreaking.Reduce(s.Item(), e.Level(), d)
	}
	if s = s.Damage(d); s.Empty() {
		p.tx.PlaySound(p.Position(), sound.ItemBreak{})
	}
	return s
}

// subtractItem subtracts d from the count of the item stack passed and returns it, if the player is in
// survival or adventure mode.
func (p *Player) subtractItem(s item.Stack, d int) item.Stack {
	if !p.GameMode().CreativeInventory() && d != 0 {
		return s.Grow(-d)
	}
	return s
}

// addNewItem adds the new item of the context passed to the inventory.
func (p *Player) addNewItem(ctx *item.UseContext) {
	if (ctx.NewItemSurvivalOnly && p.GameMode().CreativeInventory()) || ctx.NewItem.Empty() {
		return
	}
	held, left := p.HeldItems()
	if held.Empty() {
		p.SetHeldItems(ctx.NewItem, left)
		return
	}
	n, err := p.Inventory().AddItem(ctx.NewItem)
	if err != nil {
		// Not all items could be added to the inventory, so drop the rest.
		p.Drop(ctx.NewItem.Grow(ctx.NewItem.Count() - n))
	}
	if p.Dead() {
		p.dropItems()
	}
}

// canReach checks if a player can reach a position with its current range. The range depends on if the player
// is either survival or creative mode.
func (p *Player) canReach(pos mgl64.Vec3) bool {
	dist := entity.EyePosition(p).Sub(pos).Len()
	return !p.Dead() && p.GameMode().AllowsInteraction() &&
		(dist <= 8.0 || (dist <= 14.0 && p.GameMode().CreativeInventory()))
}

// Disconnect closes the player and removes it from the world.
// Disconnect, unlike Close, allows a custom message to be passed to show to the player when it is
// disconnected. The message is formatted following the rules of fmt.Sprintln without a newline at the end.
func (p *Player) Disconnect(msg ...any) {
	p.once.Do(func() {
		p.close(format(msg))
	})
}

// Close closes the player and removes it from the world.
// Close disconnects the player with a 'Connection closed.' message. Disconnect should be used to disconnect a
// player with a custom message.
func (p *Player) Close() error {
	p.once.Do(func() {
		p.close("Connection closed.")
	})
	return nil
}

// close closes the player without disconnecting it. It executes code shared by both the closing and the
// disconnecting of players.
func (p *Player) close(msg string) {
	// If the player is being disconnected while they are dead, we respawn the player
	// so that the player logic works correctly the next time they join.
	if p.Dead() && p.session() != nil {
		p.respawn(func(np *Player) {
			np.quit(msg)
		})
		return
	}
	p.quit(msg)
}

func (p *Player) quit(msg string) {
	p.h.HandleQuit(p)
	p.h = NopHandler{}

	if s := p.s; s != nil {
		s.Disconnect(msg)
		s.CloseConnection()
		return
	}
	// Only remove the player from the world if it's not attached to a session. If it is attached to a session, the
	// session will remove the player once ready.
	p.tx.RemoveEntity(p)
	_ = p.handle.Close()
}

// Data returns the player data that needs to be saved. This is used when the player
// gets disconnected and the player provider needs to save the data.
func (p *Player) Data() Config {
	p.hunger.mu.RLock()
	defer p.hunger.mu.RUnlock()
	return Config{
		Session:             p.s,
		Skin:                p.skin,
		XUID:                p.xuid,
		UUID:                p.UUID(),
		Name:                p.nameTag,
		Locale:              p.locale,
		GameMode:            p.gameMode,
		Position:            p.Position(),
		Rotation:            p.Rotation(),
		Velocity:            p.Velocity(),
		Health:              p.Health(),
		MaxHealth:           p.MaxHealth(),
		FoodTick:            p.hunger.foodTick,
		Food:                p.hunger.foodLevel,
		Exhaustion:          p.hunger.exhaustionLevel,
		Saturation:          p.hunger.saturationLevel,
		AirSupply:           p.airSupplyTicks,
		MaxAirSupply:        p.maxAirSupplyTicks,
		EnchantmentSeed:     p.enchantSeed,
		Experience:          p.experience.Experience(),
		HeldSlot:            int(*p.heldSlot),
		Inventory:           p.inv,
		OffHand:             p.offHand,
		Armour:              p.armour,
		EnderChestInventory: p.enderChest,
		FireTicks:           p.fireTicks,
		FallDistance:        p.fallDistance,
		Effects:             p.Effects(),
	}
}

// session returns the network session of the player. If it has one, it is returned. If not, a no-op session
// is returned.
func (p *Player) session() *session.Session {
	if s := p.s; s != nil {
		return s
	}
	return session.Nop
}

// useContext returns an item.UseContext initialised for a Player.
func (p *Player) useContext() *item.UseContext {
	call := func(ctx *inventory.Context, slot int, it item.Stack, f func(ctx *inventory.Context, slot int, it item.Stack)) error {
		if ctx.Cancelled() {
			return fmt.Errorf("action was cancelled")
		}
		f(ctx, slot, it)
		if ctx.Cancelled() {
			return fmt.Errorf("action was cancelled")
		}
		return nil
	}
	return &item.UseContext{
		SwapHeldWithArmour: func(i int) {
			src, dst, srcInv, dstInv := int(*p.heldSlot), i, p.inv, p.armour.Inventory()
			srcIt, _ := srcInv.Item(src)
			dstIt, _ := dstInv.Item(dst)

			ctx := event.C(inventory.Holder(p))
			_ = call(ctx, src, srcIt, srcInv.Handler().HandleTake)
			_ = call(ctx, src, dstIt, srcInv.Handler().HandlePlace)
			_ = call(ctx, dst, dstIt, dstInv.Handler().HandleTake)
			if err := call(ctx, dst, srcIt, dstInv.Handler().HandlePlace); err == nil {
				_ = srcInv.SetItem(src, dstIt)
				_ = dstInv.SetItem(dst, srcIt)
				p.PlaySound(sound.EquipItem{Item: srcIt.Item()})
			}
		},
		FirstFunc: func(comparable func(item.Stack) bool) (item.Stack, bool) {
			_, left := p.HeldItems()
			if !left.Empty() && comparable(left) {
				return left, true
			}
			inv := p.Inventory()
			s, ok := inv.FirstFunc(comparable)
			if !ok {
				return item.Stack{}, false
			}
			it, _ := inv.Item(s)
			return it, ok
		},
	}
}

// Handler returns the Handler of the player.
func (p *Player) Handler() Handler {
	return p.h
}

// broadcastItems broadcasts the items held to viewers.
func (p *Player) broadcastItems(int, item.Stack, item.Stack) {
	for _, viewer := range p.viewers() {
		viewer.ViewEntityItems(p)
	}
}

// broadcastArmour broadcasts the armour equipped to viewers.
func (p *Player) broadcastArmour(_ int, before, after item.Stack) {
	if before.Comparable(after) && before.Empty() == after.Empty() {
		// Only send armour if the type of the armour changed.
		return
	}
	for _, viewer := range p.viewers() {
		viewer.ViewEntityArmour(p)
	}
}

// viewers returns a list of all viewers of the Player.
func (p *Player) viewers() []world.Viewer {
	viewers := p.tx.Viewers(p.Position())
	var s world.Viewer = p.session()
	if slices.Index(viewers, s) == -1 && p.s != nil {
		return append(viewers, p.s)
	}
	return viewers
}

// resendBlocks resends blocks in a world.World at the cube.Pos passed and the block next to it at the cube.Face passed.
func (p *Player) resendBlocks(pos cube.Pos, faces ...cube.Face) {
	if p.session() == session.Nop {
		return
	}
	p.resendBlock(pos)
	for _, f := range faces {
		p.resendBlock(pos.Side(f))
	}
}

// resendBlock resends the block at a cube.Pos in the world.World passed.
func (p *Player) resendBlock(pos cube.Pos) {
	b := p.tx.Block(pos)
	p.session().ViewBlockUpdate(pos, b, 0)
	if _, ok := b.(world.LiquidDisplacer); ok {
		liq, _ := p.tx.Liquid(pos)
		p.session().ViewBlockUpdate(pos, liq, 1)
	}
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}
