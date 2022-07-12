package player

import (
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/entity/healing"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player/bossbar"
	"github.com/df-mc/dragonfly/server/player/chat"
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
	"math"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

// Player is an implementation of a player entity. It has methods that implement the behaviour that players
// need to play in the world.
type Player struct {
	name                                string
	uuid                                uuid.UUID
	xuid                                string
	locale                              language.Tag
	pos, vel                            atomic.Value[mgl64.Vec3]
	nameTag                             atomic.Value[string]
	scoreTag                            atomic.Value[string]
	yaw, pitch, absorptionHealth, scale atomic.Float64
	once                                sync.Once

	gameMode atomic.Value[world.GameMode]

	skin atomic.Value[skin.Skin]
	// s holds the session of the player. This field should not be used directly, but instead,
	// Player.session() should be called.
	s atomic.Value[*session.Session]
	// h holds the current Handler of the player. It may be changed at any time by calling the Handle method.
	h atomic.Value[Handler]

	inv, offHand *inventory.Inventory
	armour       *inventory.Armour
	heldSlot     *atomic.Uint32

	sneaking, sprinting, swimming, flying,
	invisible, immobile, onGround, usingItem atomic.Bool
	usingSince atomic.Int64

	fireTicks    atomic.Int64
	fallDistance atomic.Float64

	breathing         bool
	airSupplyTicks    atomic.Int64
	maxAirSupplyTicks atomic.Int64

	cooldownMu sync.Mutex
	cooldowns  map[itemHash]time.Time
	// lastTickedWorld holds the world that the player was in, in the last tick.
	lastTickedWorld *world.World

	speed      atomic.Float64
	health     *entity.HealthManager
	experience *entity.ExperienceManager
	effects    *entity.EffectManager

	lastXPPickup atomic.Value[time.Time]
	immunity     atomic.Value[time.Time]

	mc *entity.MovementComputer

	breaking          atomic.Bool
	breakingPos       atomic.Value[cube.Pos]
	lastBreakDuration time.Duration

	breakParticleCounter atomic.Uint32

	hunger *hungerManager
}

// New returns a new initialised player. A random UUID is generated for the player, so that it may be
// identified over network. You can either pass on player data you want to load or
// you can leave the data as nil to use default data.
func New(name string, skin skin.Skin, pos mgl64.Vec3) *Player {
	p := &Player{}
	*p = Player{
		inv: inventory.New(36, func(slot int, item item.Stack) {
			if slot == int(p.heldSlot.Load()) {
				p.broadcastItems(slot, item)
			}
		}),
		uuid:              uuid.New(),
		offHand:           inventory.New(1, p.broadcastItems),
		armour:            inventory.NewArmour(p.broadcastArmour),
		hunger:            newHungerManager(),
		health:            entity.NewHealthManager(),
		experience:        entity.NewExperienceManager(),
		effects:           entity.NewEffectManager(),
		gameMode:          *atomic.NewValue[world.GameMode](world.GameModeSurvival),
		h:                 *atomic.NewValue[Handler](NopHandler{}),
		name:              name,
		skin:              *atomic.NewValue(skin),
		speed:             *atomic.NewFloat64(0.1),
		nameTag:           *atomic.NewValue(name),
		heldSlot:          atomic.NewUint32(0),
		locale:            language.BritishEnglish,
		breathing:         true,
		airSupplyTicks:    *atomic.NewInt64(300),
		maxAirSupplyTicks: *atomic.NewInt64(300),
		scale:             *atomic.NewFloat64(1),
		immunity:          *atomic.NewValue(time.Now()),
		pos:               *atomic.NewValue(pos),
		cooldowns:         make(map[itemHash]time.Time),
		mc:                &entity.MovementComputer{Gravity: 0.06, Drag: 0.02, DragBeforeGravity: true},
	}
	return p
}

// NewWithSession returns a new player for a network session, so that the network session can control the
// player.
// A set of additional fields must be provided to initialise the player with the client's data, such as the
// name and the skin of the player. You can either pass on player data you want to load or
// you can leave the data as nil to use default data.
func NewWithSession(name, xuid string, uuid uuid.UUID, skin skin.Skin, s *session.Session, pos mgl64.Vec3, data *Data) *Player {
	p := New(name, skin, pos)
	p.s, p.uuid, p.xuid, p.skin = *atomic.NewValue(s), uuid, xuid, *atomic.NewValue(skin)
	p.inv, p.offHand, p.armour, p.heldSlot = s.HandleInventories()
	p.locale, _ = language.Parse(strings.Replace(s.ClientData().LanguageCode, "_", "-", 1))
	if data != nil {
		p.load(*data)
	}
	return p
}

// Name returns the username of the player. If the player is controlled by a client, it is the username of
// the client. (Typically the XBOX Live name)
func (p *Player) Name() string {
	return p.name
}

// UUID returns the UUID of the player. This UUID will remain consistent with an XBOX Live account, and will,
// unlike the name of the player, never change.
// It is therefore recommended using the UUID over the name of the player. Additionally, it is recommended to
// use the UUID over the XUID because of its standard format.
func (p *Player) UUID() uuid.UUID {
	return p.uuid
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
	return p.skin.Load()
}

// SetSkin changes the skin of the player. This skin will be visible to other players that the player
// is shown to.
func (p *Player) SetSkin(skin skin.Skin) {
	if p.Dead() {
		return
	}
	ctx := event.C()
	if p.Handler().HandleSkinChange(ctx, &skin); ctx.Cancelled() {
		p.session().ViewSkin(p)
		return
	}
	p.skin.Store(skin)
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
	p.h.Store(h)
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
	p.fallDistance.Store(0)
}

// FallDistance returns the player's fall distance.
func (p *Player) FallDistance() float64 {
	return p.fallDistance.Load()
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
	ctx := event.C()
	if p.Handler().HandleChat(ctx, &message); ctx.Cancelled() {
		return
	}
	_, _ = fmt.Fprintf(chat.Global, "<%v> %v\n", p.name, message)
}

// ExecuteCommand executes a command passed as the player. If the command could not be found, or if the usage
// was incorrect, an error message is sent to the player. This message should start with a "/" for the command to be
// recognised.
func (p *Player) ExecuteCommand(commandLine string) {
	if p.Dead() {
		return
	}
	args := strings.Split(commandLine, " ")
	command, ok := cmd.ByAlias(args[0][1:])
	if !ok {
		o := &cmd.Output{}
		o.Errorf("Unknown command: %v. Please check that the command exists and that you have permission to use it.", args[0])
		p.SendCommandOutput(o)
		return
	}
	ctx := event.C()
	if p.Handler().HandleCommandExecution(ctx, command, args[1:]); ctx.Cancelled() {
		return
	}
	command.Execute(strings.Join(args[1:], " "), p)
}

// Transfer transfers the player to a server at the address passed. If the address could not be resolved, an
// error is returned. If it is returned, the player is closed and transferred to the server.
func (p *Player) Transfer(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	ctx := event.C()
	if p.Handler().HandleTransfer(ctx, addr); ctx.Cancelled() {
		return nil
	}
	p.session().Transfer(addr.IP, addr.Port)
	return nil
}

// SendCommandOutput sends the output of a command to the player.
func (p *Player) SendCommandOutput(output *cmd.Output) {
	p.session().SendCommandOutput(output)
}

// SendForm sends a form to the player for the client to fill out. Once the client fills it out, the Submit
// method of the form will be called.
// Note that the client may also close the form instead of filling it out, which will result in the form not
// having its Submit method called at all. Forms should never depend on the player actually filling out the
// form.
func (p *Player) SendForm(f form.Form) {
	p.session().SendForm(f)
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
	p.nameTag.Store(name)
	p.updateState()
}

// NameTag returns the current name tag of the Player as shown in-game. It can be changed using SetNameTag.
func (p *Player) NameTag() string {
	return p.nameTag.Load()
}

// SetScoreTag changes the score tag displayed over the player in-game. The score tag is displayed under the player's
// name tag.
func (p *Player) SetScoreTag(a ...any) {
	p.scoreTag.Store(format(a))
	p.updateState()
}

// ScoreTag returns the current score tag of the player. It can be changed using SetScoreTag and by default is empty.
func (p *Player) ScoreTag() string {
	return p.scoreTag.Load()
}

// SetSpeed sets the speed of the player. The value passed is the blocks/tick speed that the player will then
// obtain.
func (p *Player) SetSpeed(speed float64) {
	p.speed.Store(speed)
	p.session().SendSpeed(speed)
}

// Speed returns the speed of the player, returning a value that indicates the blocks/tick speed. The default
// speed of a player is 0.1.
func (p *Player) Speed() float64 {
	return p.speed.Load()
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
	p.session().SendHealth(p.health)
}

// addHealth adds health to the player's current health.
func (p *Player) addHealth(health float64) {
	p.health.AddHealth(health)
	p.session().SendHealth(p.health)
}

// Heal heals the entity for a given amount of health. The source passed represents the cause of the
// healing, for example healing.SourceFood if the entity healed by having a full food bar. If the health
// added to the original health exceeds the entity's max health, Heal will not add the full amount.
// If the health passed is negative, Heal will not do anything.
func (p *Player) Heal(health float64, source healing.Source) {
	if p.Dead() || health < 0 || !p.GameMode().AllowsTakingDamage() {
		return
	}
	ctx := event.C()
	if p.Handler().HandleHeal(ctx, &health, source); ctx.Cancelled() {
		return
	}
	p.addHealth(health)
}

// updateFallState is called to update the entities falling state.
func (p *Player) updateFallState(distanceThisTick float64) {
	fallDistance := p.fallDistance.Load()
	if p.OnGround() {
		if fallDistance > 0 {
			p.fall(fallDistance)
			p.ResetFallDistance()
		}
	} else if distanceThisTick < fallDistance {
		p.fallDistance.Sub(distanceThisTick)
	} else {
		p.ResetFallDistance()
	}
}

// fall is called when a falling entity hits the ground.
func (p *Player) fall(distance float64) {
	var (
		w   = p.World()
		pos = cube.PosFromVec3(p.Position())
		b   = w.Block(pos)
	)
	if len(b.Model().BBox(pos, w)) == 0 {
		pos = pos.Sub(cube.Pos{0, 1})
		b = w.Block(pos)
	}
	if h, ok := b.(block.EntityLander); ok {
		h.EntityLand(pos, w, p, &distance)
	}
	dmg := distance - 3
	if boost, ok := p.Effect(effect.JumpBoost{}); ok {
		dmg -= float64(boost.Level())
	}
	if dmg < 0.5 {
		return
	}
	p.Hurt(math.Ceil(dmg), damage.SourceFall{})
}

// Hurt hurts the player for a given amount of damage. The source passed represents the cause of the damage,
// for example damage.SourceEntityAttack if the player is attacked by another entity.
// If the final damage exceeds the health that the player currently has, the player is killed and will have to
// respawn.
// If the damage passed is negative, Hurt will not do anything.
// Hurt returns the final damage dealt to the Player and if the Player was vulnerable to this kind of damage.
func (p *Player) Hurt(dmg float64, source damage.Source) (float64, bool) {
	if p.Dead() || !p.GameMode().AllowsTakingDamage() {
		return 0, false
	}

	fireSource := source == damage.SourceFire{} || source == damage.SourceFireTick{} || source == damage.SourceLava{}
	if _, ok := p.Effect(effect.FireResistance{}); ok && fireSource {
		return 0, false
	}

	immunity := time.Second / 2
	ctx := event.C()
	if p.Handler().HandleHurt(ctx, &dmg, &immunity, source); ctx.Cancelled() {
		return 0, false
	}
	if dmg < 0 {
		return 0, true
	}

	totalDamage := p.FinalDamageFrom(dmg, source)
	damageLeft := totalDamage

	if a := p.Absorption(); a > 0 {
		if damageLeft > a {
			damageLeft -= a
			p.SetAbsorption(0)
			p.effects.Remove(effect.Absorption{}, p)
		} else {
			p.SetAbsorption(a - damageLeft)
			damageLeft = 0
		}
	}
	p.addHealth(-damageLeft)

	if source.ReducedByArmour() {
		p.Exhaust(0.1)

		damageToArmour := int(math.Max(math.Floor(dmg/4), 1))
		for slot, it := range p.armour.Slots() {
			if _, ok := it.Item().(item.Durable); ok {
				_ = p.armour.Inventory().SetItem(slot, p.damageItem(it, damageToArmour))
			}
		}
	}

	w, pos := p.World(), p.Position()
	for _, viewer := range p.viewers() {
		viewer.ViewEntityAction(p, entity.HurtAction{})
	}
	if fireSource {
		w.PlaySound(pos, sound.Burning{})
	}
	if _, ok := source.(damage.SourceDrowning); ok {
		w.PlaySound(pos, sound.Drowning{})
	}

	p.immunity.Store(time.Now().Add(immunity))
	if p.Dead() {
		p.kill(source)
	}
	return totalDamage, true
}

// FinalDamageFrom resolves the final damage received by the player if it is attacked by the source passed
// with the damage passed. FinalDamageFrom takes into account things such as the armour worn and the
// enchantments on the individual pieces.
// The damage returned will be at the least 0.
func (p *Player) FinalDamageFrom(dmg float64, src damage.Source) float64 {
	if src.ReducedByArmour() {
		defencePoints := 0.0
		toughness := 0.0
		for _, it := range p.armour.Items() {
			if a, ok := it.Item().(item.Armour); ok {
				defencePoints += a.DefencePoints()
				toughness += a.Toughness()
			}
		}
		// Armour in Bedrock edition reduces the damage taken by 4% for each effective armour point. Effective
		// armour point decreases as damage increases, with 1 point lost for every 2 HP of damage. The defense
		// reduction is decreased by the toughness armor value. Effective armour points will at minimum be 20% of
		// armour points.
		dmg -= dmg * 0.04 * math.Max(defencePoints*0.2, defencePoints-dmg/(2+toughness/4))
	}
	if res, ok := p.Effect(effect.Resistance{}); ok {
		dmg *= effect.Resistance{}.Multiplier(src, res.Level())
	}

	f := p.enchantmentProtectionFactor(src)
	if f > 25 {
		f = 25
	}
	m := math.Ceil(float64(f) * (float64(rand.Intn(100-50)+50) / 100.0))
	if m > 20 {
		m = 20
	}

	dmg -= -dmg * m * 0.04
	return math.Max(dmg, 0)
}

// protectionEnchantment represents an enchantment that can protect the player from damage.
type protectionEnchantment interface {
	Affects(damage.Source) bool
	Modifier() float64
}

// enchantmentProtectionFactor returns the combined enchantment protection factor the player inhibits spanning each
// armour piece.
func (p *Player) enchantmentProtectionFactor(src damage.Source) (f int) {
	for _, it := range p.armour.Items() {
		for _, e := range it.Enchantments() {
			if p, ok := e.Type().(protectionEnchantment); ok && p.Affects(src) {
				lvl := e.Level()
				f += int(math.Floor(float64(6+lvl*lvl) * p.Modifier() / 3))
			}
		}
	}
	return f
}

// highestArmourEnchantmentLevel returns the highest level of the enchantment passed based spanning each armour piece.
func (p *Player) highestArmourEnchantmentLevel(enchant item.EnchantmentType) (t int) {
	for _, it := range p.armour.Items() {
		if e, ok := it.Enchantment(enchant); ok {
			if e.Level() > t {
				t = e.Level()
			}
		}
	}
	return t
}

// SetAbsorption sets the absorption health of a player. This extra health shows as golden hearts and do not
// actually increase the maximum health. Once the hearts are lost, they will not regenerate.
// Nothing happens if a negative number is passed.
func (p *Player) SetAbsorption(health float64) {
	health = math.Max(health, 0)
	p.absorptionHealth.Store(health)
	p.session().SendAbsorption(health)
}

// Absorption returns the absorption health that the player has.
func (p *Player) Absorption() float64 {
	return p.absorptionHealth.Load()
}

// KnockBack knocks the player back with a given force and height. A source is passed which indicates the
// source of the velocity, typically the position of an attacking entity. The source is used to calculate the
// direction which the entity should be knocked back in.
func (p *Player) KnockBack(src mgl64.Vec3, force, height float64) {
	if p.Dead() || !p.GameMode().AllowsTakingDamage() {
		return
	}
	velocity := p.Position().Sub(src)
	velocity[1] = 0

	velocity = velocity.Normalize().Mul(force)
	velocity[1] = height

	resistance := 0.0
	for _, i := range p.armour.Items() {
		if a, ok := i.Item().(item.Armour); ok {
			resistance += a.KnockBackResistance()
		}
	}

	p.SetVelocity(velocity.Mul(1 - resistance))
}

// AttackImmune checks if the player is currently immune to entity attacks, meaning it was recently attacked.
func (p *Player) AttackImmune() bool {
	return p.immunity.Load().After(time.Now())
}

// AttackImmunity returns the duration the player is immune to entity attacks.
func (p *Player) AttackImmunity() time.Duration {
	return time.Until(p.immunity.Load())
}

// SetAttackImmunity sets the duration the player is immune to entity attacks.
func (p *Player) SetAttackImmunity(d time.Duration) {
	p.immunity.Store(time.Now().Add(d))
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
	p.hunger.mu.RLock()
	defer p.hunger.mu.RUnlock()
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
	if !p.GameMode().AllowsTakingDamage() || p.World().Difficulty().FoodRegenerates() {
		return
	}
	before := p.hunger.Food()
	p.hunger.exhaust(points)
	if after := p.hunger.Food(); before != after {
		// Temporarily set the food level back so that it hasn't yet changed once the event is handled.
		p.hunger.SetFood(before)

		ctx := event.C()
		if p.Handler().HandleFoodLoss(ctx, before, after); ctx.Cancelled() {
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
	return p.Health() <= 0
}

// kill kills the player, clearing its inventories and resetting it to its base state.
func (p *Player) kill(src damage.Source) {
	for _, viewer := range p.viewers() {
		viewer.ViewEntityAction(p, entity.DeathAction{})
	}

	p.addHealth(-p.MaxHealth())

	p.Handler().HandleDeath(src)
	p.StopSneaking()
	p.StopSprinting()

	w := p.World()
	p.dropContents()

	for _, e := range p.Effects() {
		p.RemoveEffect(e.Type())
	}

	// Wait a little before removing the entity. The client displays a death animation while the player is dying.
	time.AfterFunc(time.Millisecond*1100, func() {
		if p.session() == session.Nop {
			_ = p.Close()
			return
		}
		if p.Dead() {
			p.SetInvisible()
			// We have an actual client connected to this player: We change its position server side so that in
			// the future, the client won't respawn on the death location when disconnecting. The client should
			// not see the movement itself yet, though.
			p.pos.Store(w.Spawn().Vec3())
		}
	})
}

// dropContents drops all items and experience of the Player on the ground in random directions.
func (p *Player) dropContents() {
	w, pos := p.World(), p.Position()
	for _, orb := range entity.NewExperienceOrbs(pos, int(math.Min(float64(p.experience.Level()*7), 100))) {
		orb.SetVelocity(mgl64.Vec3{(rand.Float64()*0.2 - 0.1) * 2, rand.Float64() * 0.4, (rand.Float64()*0.2 - 0.1) * 2})
		w.AddEntity(orb)
	}
	for _, it := range append(p.inv.Items(), append(p.armour.Items(), p.offHand.Items()...)...) {
		ent := entity.NewItem(it, pos)
		ent.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
		w.AddEntity(ent)
	}
	p.inv.Clear()
	p.armour.Clear()
	p.offHand.Clear()
	p.experience.Reset()
	p.session().SendExperience(p.experience)
}

// Respawn spawns the player after it dies, so that its health is replenished and it is spawned in the world
// again. Nothing will happen if the player does not have a session connected to it.
func (p *Player) Respawn() {
	w := p.World()
	if !p.Dead() || w == nil || p.session() == session.Nop {
		return
	}
	p.addHealth(p.MaxHealth())
	p.hunger.Reset()
	p.sendFood()
	p.Extinguish()
	p.ResetFallDistance()

	// We can use the principle here that returning through a portal of a specific dimension inside that dimension will
	// always bring us back to the overworld.
	w = w.PortalDestination(w.Dimension())
	pos := w.PlayerSpawn(p.UUID()).Vec3Middle()

	p.Handler().HandleRespawn(&pos, &w)

	w.AddEntity(p)
	p.Teleport(pos)
	p.session().SendRespawn(pos)

	p.SetVisible()
}

// StartSprinting makes a player start sprinting, increasing the speed of the player by 30% and making
// particles show up under the feet. The player will only start sprinting if its food level is high enough.
// If the player is sneaking when calling StartSprinting, it is stopped from sneaking.
func (p *Player) StartSprinting() {
	if !p.hunger.canSprint() && p.GameMode().AllowsTakingDamage() {
		return
	}
	ctx := event.C()
	if p.Handler().HandleToggleSprint(ctx, true); ctx.Cancelled() {
		return
	}
	if !p.sprinting.CAS(false, true) {
		return
	}
	p.StopSneaking()
	p.SetSpeed(p.Speed() * 1.3)

	p.updateState()
}

// Sprinting checks if the player is currently sprinting.
func (p *Player) Sprinting() bool {
	return p.sprinting.Load()
}

// StopSprinting makes a player stop sprinting, setting back the speed of the player to its original value.
func (p *Player) StopSprinting() {
	ctx := event.C()
	if p.Handler().HandleToggleSprint(ctx, false); ctx.Cancelled() {
		return
	}
	if !p.sprinting.CAS(true, false) {
		return
	}
	p.SetSpeed(p.Speed() / 1.3)

	p.updateState()
}

// StartSneaking makes a player start sneaking. If the player is already sneaking, StartSneaking will not do
// anything.
// If the player is sprinting while StartSneaking is called, the sprinting is stopped.
func (p *Player) StartSneaking() {
	ctx := event.C()
	if p.Handler().HandleToggleSneak(ctx, true); ctx.Cancelled() {
		return
	}
	if !p.sneaking.CAS(false, true) {
		return
	}
	if !p.Flying() {
		p.StopSprinting()
	}
	p.updateState()
}

// Sneaking checks if the player is currently sneaking.
func (p *Player) Sneaking() bool {
	return p.sneaking.Load()
}

// StopSneaking makes a player stop sneaking if it currently is. If the player is not sneaking, StopSneaking
// will not do anything.
func (p *Player) StopSneaking() {
	ctx := event.C()
	if p.Handler().HandleToggleSneak(ctx, false); ctx.Cancelled() {
		return
	}
	if !p.sneaking.CAS(true, false) {
		return
	}
	p.updateState()
}

// StartSwimming makes the player start swimming if it is not currently doing so. If the player is sneaking
// while StartSwimming is called, the sneaking is stopped.
func (p *Player) StartSwimming() {
	if !p.swimming.CAS(false, true) {
		return
	}
	p.StopSneaking()
	p.updateState()
}

// Swimming checks if the player is currently swimming.
func (p *Player) Swimming() bool {
	return p.swimming.Load()
}

// StopSwimming makes the player stop swimming if it is currently doing so.
func (p *Player) StopSwimming() {
	if !p.swimming.CAS(true, false) {
		return
	}
	p.updateState()
}

// StartFlying makes the player start flying if they aren't already. It requires the player to be in a gamemode which
// allows flying.
func (p *Player) StartFlying() {
	if !p.GameMode().AllowsFlying() || !p.flying.CAS(false, true) {
		return
	}
	p.session().SendGameMode(p.GameMode())
}

// Flying checks if the player is currently flying.
func (p *Player) Flying() bool {
	return p.flying.Load()
}

// StopFlying makes the player stop flying if it currently is.
func (p *Player) StopFlying() {
	if !p.flying.CAS(true, false) {
		return
	}
	p.session().SendGameMode(p.GameMode())
}

// Jump makes the player jump if they are on ground. It exhausts the player by 0.05 food points, an additional 0.15
// is exhausted if the player is sprint jumping.
func (p *Player) Jump() {
	if p.Dead() {
		return
	}

	p.Handler().HandleJump()
	if p.OnGround() {
		jumpVel := 0.42
		if e, ok := p.Effect(effect.JumpBoost{}); ok {
			jumpVel = float64(e.Level()) / 10
		}
		p.vel.Store(mgl64.Vec3{0, jumpVel})
	}
	if p.Sprinting() {
		p.Exhaust(0.2)
	} else {
		p.Exhaust(0.05)
	}
}

// SetInvisible sets the player invisible, so that other players will not be able to see it.
func (p *Player) SetInvisible() {
	if !p.invisible.CAS(false, true) {
		return
	}
	p.updateState()
}

// SetVisible sets the player visible again, so that other players can see it again. If the player was already
// visible, or if the player is in spectator mode, nothing happens.
func (p *Player) SetVisible() {
	if !p.GameMode().Visible() {
		return
	}
	if _, ok := p.Effect(effect.Invisibility{}); ok {
		return
	}
	if !p.invisible.CAS(true, false) {
		return
	}
	p.updateState()
}

// Invisible checks if the Player is currently invisible.
func (p *Player) Invisible() bool {
	return p.invisible.Load()
}

// SetImmobile prevents the player from moving around, but still allows them to look around.
func (p *Player) SetImmobile() {
	if !p.immobile.CAS(false, true) {
		return
	}
	p.updateState()
}

// SetMobile allows the player to freely move around again after being immobile.
func (p *Player) SetMobile() {
	if !p.immobile.CAS(true, false) {
		return
	}
	p.updateState()
}

// Immobile checks if the Player is currently immobile.
func (p *Player) Immobile() bool {
	return p.immobile.Load()
}

// FireProof checks if the Player is currently fireproof. True is returned if the player has a FireResistance effect or
// if it is in creative mode.
func (p *Player) FireProof() bool {
	if _, ok := p.Effect(effect.FireResistance{}); ok {
		return true
	}
	return !p.GameMode().AllowsTakingDamage()
}

// OnFireDuration ...
func (p *Player) OnFireDuration() time.Duration {
	return time.Duration(p.fireTicks.Load()) * time.Second / 20
}

// SetOnFire ...
func (p *Player) SetOnFire(duration time.Duration) {
	ticks := int64(duration.Seconds() * 20)
	if level := p.highestArmourEnchantmentLevel(enchantment.FireProtection{}); level > 0 {
		ticks -= int64(math.Floor(float64(ticks) * float64(level) * 0.15))
	}
	p.fireTicks.Store(ticks)
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
	mainHand, _ = p.inv.Item(int(p.heldSlot.Load()))
	return mainHand, offHand
}

// SetHeldItems sets items to the main hand and the off-hand of the player. The Stacks passed may be empty
// (Stack.Empty()) to clear the held item.
func (p *Player) SetHeldItems(mainHand, offHand item.Stack) {
	_ = p.inv.SetItem(int(p.heldSlot.Load()), mainHand)
	_ = p.offHand.SetItem(0, offHand)
}

// SetGameMode sets the game mode of a player. The game mode specifies the way that the player can interact
// with the world that it is in.
func (p *Player) SetGameMode(mode world.GameMode) {
	previous := p.gameMode.Swap(mode)
	p.session().SendGameMode(mode)

	if !mode.AllowsFlying() {
		p.StopFlying()
	}
	if !mode.Visible() {
		p.SetInvisible()
	} else if !previous.Visible() {
		p.SetVisible()
	}
}

// GameMode returns the current game mode assigned to the player. If not changed, the game mode returned will
// be the same as that of the world that the player spawns in.
// The game mode may be changed using Player.SetGameMode().
func (p *Player) GameMode() world.GameMode {
	return p.gameMode.Load()
}

// itemHash is used as a hash for a world.Item.
type itemHash struct {
	// Name is the name of the item.
	Name string
	// Meta is the item's metadata value.
	Meta int16
}

// hashFromItem returns an item hash from an item.
func hashFromItem(item world.Item) itemHash {
	name, meta := item.EncodeItem()
	return itemHash{
		Name: name,
		Meta: meta,
	}
}

// HasCooldown returns true if the item passed has an active cooldown, meaning it currently cannot be used again. If the
// world.Item passed is nil, HasCooldown always returns false.
func (p *Player) HasCooldown(item world.Item) bool {
	if item == nil {
		return false
	}
	p.cooldownMu.Lock()
	defer p.cooldownMu.Unlock()

	hash := hashFromItem(item)
	otherTime, ok := p.cooldowns[hash]
	if !ok {
		return false
	}
	if time.Now().After(otherTime) {
		delete(p.cooldowns, hash)
		return false
	}
	return true
}

// SetCooldown sets a cooldown for an item. If the world.Item passed is nil, nothing happens.
func (p *Player) SetCooldown(item world.Item, cooldown time.Duration) {
	if item == nil {
		return
	}
	p.cooldownMu.Lock()
	defer p.cooldownMu.Unlock()

	p.cooldowns[hashFromItem(item)] = time.Now().Add(cooldown)
}

// UseItem uses the item currently held in the player's main hand in the air. Generally, nothing happens,
// unless the held item implements the item.Usable interface, in which case it will be activated.
// This generally happens for items such as throwable items like snowballs.
func (p *Player) UseItem() {
	var (
		i, left = p.HeldItems()
		w       = p.World()
		ctx     = event.C()
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

	switch usable := it.(type) {
	case item.Releasable:
		if !p.canRelease() {
			return
		}
		p.usingSince.Store(time.Now().UnixNano())
		p.usingItem.Store(true)
		p.updateState()
	case item.Usable:
		useCtx := p.useContext()
		if !usable.Use(w, p, useCtx) {
			return
		}
		// We only swing the player's arm if the item held actually does something. If it doesn't, there is no
		// reason to swing the arm.
		p.SwingArm()
		p.SetHeldItems(p.subtractItem(p.damageItem(i, useCtx.Damage), useCtx.CountSub), left)
		p.addNewItem(useCtx)
	case item.Consumable:
		if !usable.AlwaysConsumable() && p.GameMode().AllowsTakingDamage() && p.Food() >= 20 {
			// The item.Consumable is not always consumable, the player is not in creative mode and the
			// food bar is filled: The item cannot be consumed.
			p.ReleaseItem()
			return
		}
		if p.usingItem.CAS(false, true) {
			// Consumable starts being consumed: Set the start timestamp and update the using state to viewers.
			p.usingSince.Store(time.Now().UnixNano())
			p.updateState()
			return
		}
		// The player is currently using the item held. This is a signal the item was consumed, so we
		// consume it and start using it again.
		p.ReleaseItem()
		if duration := p.useDuration(); duration < usable.ConsumeDuration() {
			// The required duration for consuming this item was not met, so we don't consume it.
			return
		}

		ctx = event.C()
		if p.Handler().HandleItemConsume(ctx, i); ctx.Cancelled() {
			// Consuming was cancelled, but the client will continue consuming the next item.
			p.usingSince.Store(time.Now().UnixNano())
			return
		}
		p.SetHeldItems(p.subtractItem(i, 1), left)

		useCtx := p.useContext()
		useCtx.NewItem = usable.Consume(w, p)
		p.addNewItem(useCtx)
		w.PlaySound(p.Position().Add(mgl64.Vec3{0, 1.5}), sound.Burp{})
	}
}

// ReleaseItem makes the Player release the item it is currently using. This is only applicable for items that
// implement the item.Releasable interface.
// If the Player is not currently using any item, ReleaseItem returns immediately.
// ReleaseItem either aborts the using of the item or finished it, depending on the time that elapsed since
// the item started being used.
func (p *Player) ReleaseItem() {
	if !p.usingItem.CAS(true, false) || !p.canRelease() || !p.GameMode().AllowsInteraction() {
		return
	}
	ctx := p.useContext()
	i, _ := p.HeldItems()
	i.Item().(item.Releasable).Release(p, p.useDuration(), ctx)

	p.handleUseContext(ctx)
	p.updateState()
}

// canRelease returns whether the player can release the item currently held in the main hand.
func (p *Player) canRelease() bool {
	held, _ := p.HeldItems()
	releasable, ok := held.Item().(item.Releasable)
	if !ok {
		return false
	}
	if p.GameMode().CreativeInventory() {
		return true
	}
	for _, req := range releasable.Requirements() {
		_, found := p.Inventory().FirstFunc(func(stack item.Stack) bool {
			name, _ := stack.Item().EncodeItem()
			otherName, _ := req.Item().EncodeItem()
			return name == otherName
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
		_ = p.Inventory().RemoveItem(it)
	}
}

// useDuration returns the duration the player has been using the item in the main hand.
func (p *Player) useDuration() time.Duration {
	return time.Duration(time.Now().UnixNano()-p.usingSince.Load()) + time.Second/20
}

// UsingItem checks if the Player is currently using an item. True is returned if the Player is currently eating an
// item or using it over a longer duration such as when using a bow.
func (p *Player) UsingItem() bool {
	return p.usingItem.Load()
}

// UseItemOnBlock uses the item held in the main hand of the player on a block at the position passed. The
// player is assumed to have clicked the face passed with the relative click position clickPos.
// If the item could not be used successfully, for example when the position is out of range, the method
// returns immediately.
// UseItemOnBlock does nothing if the block at the cube.Pos passed is of the type block.Air.
func (p *Player) UseItemOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	w := p.World()
	if _, ok := w.Block(pos).(block.Air); ok || !p.canReach(pos.Vec3Centre()) {
		// The client used its item on a block that does not exist server-side or one it couldn't reach. Stop trying
		// to use the item immediately.
		p.resendBlocks(pos, w, face)
		return
	}
	ctx := event.C()
	if p.Handler().HandleItemUseOnBlock(ctx, pos, face, clickPos); ctx.Cancelled() {
		p.resendBlocks(pos, w, face)
		return
	}
	i, left := p.HeldItems()
	b := w.Block(pos)
	if act, ok := b.(block.Activatable); ok {
		// If a player is sneaking, it will not activate the block clicked, unless it is not holding any
		// items, in which case the block will be activated as usual.
		if !p.Sneaking() || i.Empty() {
			p.SwingArm()
			// The block was activated: Blocks such as doors must always have precedence over the item being
			// used.
			if act.Activate(pos, face, p.World(), p) {
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
		if !ib.UseOnBlock(pos, face, clickPos, p.World(), p, useCtx) {
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
		if replaceable, ok := w.Block(replacedPos).(block.Replaceable); !ok || !replaceable.ReplaceableBy(ib) || replacedPos.OutOfBounds(w.Range()) {
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
func (p *Player) UseItemOnEntity(e world.Entity) {
	if !p.canReach(e.Position()) {
		return
	}
	ctx := event.C()
	if p.Handler().HandleItemUseOnEntity(ctx, e); ctx.Cancelled() {
		return
	}
	i, left := p.HeldItems()
	usable, ok := i.Item().(item.UsableOnEntity)
	if !ok {
		return
	}
	useCtx := p.useContext()
	if !usable.UseOnEntity(e, e.World(), p, useCtx) {
		return
	}
	p.SwingArm()
	p.SetHeldItems(p.subtractItem(p.damageItem(i, useCtx.Damage), useCtx.CountSub), left)
	p.addNewItem(useCtx)
}

// AttackEntity uses the item held in the main hand of the player to attack the entity passed, provided it is
// within range of the player.
// The damage dealt to the entity will depend on the item held by the player and any effects the player may
// have.
// If the player cannot reach the entity at its position, the method returns immediately.
func (p *Player) AttackEntity(e world.Entity) {
	if !p.canReach(e.Position()) {
		return
	}
	var (
		force, height  = 0.45, 0.3608
		_, slowFalling = p.Effect(effect.SlowFalling{})
		_, blind       = p.Effect(effect.Blindness{})
		critical       = !p.Sprinting() && !p.Flying() && p.FallDistance() > 0 && !slowFalling && !blind
	)

	ctx := event.C()
	if p.Handler().HandleAttackEntity(ctx, e, &force, &height, &critical); ctx.Cancelled() {
		return
	}
	p.SwingArm()

	i, _ := p.HeldItems()
	living, ok := e.(entity.Living)
	if !ok || living.AttackImmune() {
		return
	}

	dmg := i.AttackDamage()
	if strength, ok := p.Effect(effect.Strength{}); ok {
		dmg += dmg * effect.Strength{}.Multiplier(strength.Level())
	}
	if weakness, ok := p.Effect(effect.Weakness{}); ok {
		dmg -= dmg * effect.Weakness{}.Multiplier(weakness.Level())
	}
	if s, ok := i.Enchantment(enchantment.Sharpness{}); ok {
		dmg += (enchantment.Sharpness{}).Addend(s.Level())
	}
	if critical {
		dmg *= 1.5
	}

	n, vulnerable := living.Hurt(dmg, damage.SourceEntityAttack{Attacker: p})
	i, left := p.HeldItems()

	p.World().PlaySound(entity.EyePosition(e), sound.Attack{Damage: !mgl64.FloatEqual(n, 0)})
	if !vulnerable {
		return
	}
	if critical {
		for _, v := range p.World().Viewers(living.Position()) {
			v.ViewEntityAction(living, entity.CriticalHitAction{})
		}
	}

	p.Exhaust(0.1)

	if k, ok := i.Enchantment(enchantment.KnockBack{}); ok {
		inc := (enchantment.KnockBack{}).Force(k.Level())
		force += inc
		height += inc
	}
	living.KnockBack(p.Position(), force, height)

	if f, ok := i.Enchantment(enchantment.FireAspect{}); ok {
		if flammable, ok := living.(entity.Flammable); ok {
			flammable.SetOnFire((enchantment.FireAspect{}).Duration(f.Level()))
		}
	}

	if durable, ok := i.Item().(item.Durable); ok {
		p.SetHeldItems(p.damageItem(i, durable.DurabilityInfo().AttackDurability), left)
	}
}

// StartBreaking makes the player start breaking the block at the position passed using the item currently
// held in its main hand.
// If no block is present at the position, or if the block is out of range, StartBreaking will return
// immediately and the block will not be broken. StartBreaking will stop the breaking of any block that the
// player might be breaking before this method is called.
func (p *Player) StartBreaking(pos cube.Pos, face cube.Face) {
	p.AbortBreaking()
	w := p.World()
	if _, air := w.Block(pos).(block.Air); air || !p.canReach(pos.Vec3Centre()) {
		// The block was either out of range or air, so it can't be broken by the player.
		return
	}
	if _, ok := w.Block(pos.Side(face)).(block.Fire); ok {
		// TODO: Add a way to cancel fire extinguishing. This is currently not possible to handle.
		w.SetBlock(pos.Side(face), nil, nil)
		w.PlaySound(pos.Vec3(), sound.FireExtinguish{})
		return
	}

	held, _ := p.HeldItems()
	if _, ok := held.Item().(item.Sword); ok && p.GameMode().CreativeInventory() {
		// Can't break blocks with a sword in creative mode.
		return
	}
	// Note: We intentionally store this regardless of whether the breaking proceeds, so that we
	// can resend the block to the client when it tries to break the block regardless.
	p.breakingPos.Store(pos)

	ctx := event.C()
	if p.Handler().HandleStartBreak(ctx, pos); ctx.Cancelled() {
		return
	}
	if punchable, ok := w.Block(pos).(block.Punchable); ok {
		punchable.Punch(pos, face, w, p)
	}

	p.breaking.Store(true)
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
	w := p.World()
	breakTime := block.BreakDuration(w.Block(pos), held)
	if !p.OnGround() {
		breakTime *= 5
	}
	if _, ok := p.Armour().Helmet().Enchantment(enchantment.AquaAffinity{}); p.insideOfWater(w) && !ok {
		breakTime *= 5
	}
	for _, e := range p.Effects() {
		lvl := e.Level()
		switch v := e.Type().(type) {
		case effect.Haste:
			breakTime = time.Duration(float64(breakTime) * v.Multiplier(lvl))
		case effect.MiningFatigue:
			breakTime = time.Duration(float64(breakTime) * v.Multiplier(lvl))
		case effect.ConduitPower:
			breakTime = time.Duration(float64(breakTime) * v.Multiplier(lvl))
		}
	}
	return breakTime
}

// FinishBreaking makes the player finish breaking the block it is currently breaking, or returns immediately
// if the player isn't breaking anything.
// FinishBreaking will stop the animation and break the block.
func (p *Player) FinishBreaking() {
	pos := p.breakingPos.Load()
	if !p.breaking.Load() {
		p.resendBlock(pos, p.World())
		return
	}
	p.AbortBreaking()
	p.BreakBlock(pos)
}

// AbortBreaking makes the player stop breaking the block it is currently breaking, or returns immediately
// if the player isn't breaking anything.
// Unlike FinishBreaking, AbortBreaking does not stop the animation.
func (p *Player) AbortBreaking() {
	if !p.breaking.CAS(true, false) {
		return
	}
	p.breakParticleCounter.Store(0)
	pos := p.breakingPos.Load()
	for _, viewer := range p.viewers() {
		viewer.ViewBlockAction(pos, block.StopCrackAction{})
	}
}

// ContinueBreaking makes the player continue breaking the block it started breaking after a call to
// Player.StartBreaking().
// The face passed is used to display particles on the side of the block broken.
func (p *Player) ContinueBreaking(face cube.Face) {
	if !p.breaking.Load() {
		return
	}
	pos := p.breakingPos.Load()

	p.SwingArm()

	w := p.World()
	b := w.Block(pos)
	w.AddParticle(pos.Vec3(), particle.PunchBlock{Block: b, Face: face})

	if p.breakParticleCounter.Add(1)%5 == 0 {
		// We send this sound only every so often. Vanilla doesn't send it every tick while breaking
		// either. Every 5 ticks seems accurate.
		w.PlaySound(pos.Vec3(), sound.BlockBreaking{Block: w.Block(pos)})
	}
	breakTime := p.breakTime(pos)
	if breakTime != p.lastBreakDuration {
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
	if !p.placeBlock(pos, b, ctx.IgnoreBBox) {
		return
	}
	if ctx != nil {
		ctx.CountSub++
	}
}

// placeBlock makes the player place the block passed at the position passed, granted it is within the range
// of the player. A bool is returned indicating if a block was placed successfully.
func (p *Player) placeBlock(pos cube.Pos, b world.Block, ignoreBBox bool) bool {
	w := p.World()
	if !p.canReach(pos.Vec3Centre()) || !p.GameMode().AllowsEditing() {
		p.resendBlocks(pos, w, cube.Faces()...)
		return false
	}
	if !ignoreBBox && p.obstructedPos(pos, b) {
		p.resendBlocks(pos, w, cube.Faces()...)
		return false
	}

	ctx := event.C()
	if p.Handler().HandleBlockPlace(ctx, pos, b); ctx.Cancelled() {
		p.resendBlocks(pos, w, cube.Faces()...)
		return false
	}
	w.SetBlock(pos, b, nil)
	w.PlaySound(pos.Vec3(), sound.BlockPlace{Block: b})
	p.SwingArm()
	return true
}

// obstructedPos checks if the position passed is obstructed if the block passed is attempted to be placed.
// The function returns true if there is an entity in the way that could prevent the block from being placed.
func (p *Player) obstructedPos(pos cube.Pos, b world.Block) bool {
	w := p.World()
	blockBoxes := b.Model().BBox(pos, w)
	for i, box := range blockBoxes {
		blockBoxes[i] = box.Translate(pos.Vec3())
	}

	around := w.EntitiesWithin(cube.Box(-3, -3, -3, 3, 3, 3).Translate(pos.Vec3()), nil)
	for _, e := range around {
		if _, ok := e.(*entity.Item); ok {
			// Placing blocks inside item entities is fine.
			continue
		}
		if _, ok := e.(*entity.Arrow); ok {
			// Placing blocks inside arrow entities is fine.
			continue
		}
		if cube.AnyIntersections(blockBoxes, e.BBox().Translate(e.Position())) {
			return true
		}
	}
	return false
}

// BreakBlock makes the player break a block in the world at a position passed. If the player is unable to
// reach the block passed, the method returns immediately.
func (p *Player) BreakBlock(pos cube.Pos) {
	w := p.World()
	b := w.Block(pos)
	if _, air := b.(block.Air); air {
		// Don't do anything if the position broken is already air.
		return
	}
	if !p.canReach(pos.Vec3Centre()) || !p.GameMode().AllowsEditing() {
		p.resendBlocks(pos, w)
		return
	}
	if _, breakable := b.(block.Breakable); !breakable && !p.GameMode().CreativeInventory() {
		p.resendBlocks(pos, w)
		return
	}
	held, _ := p.HeldItems()
	drops := p.drops(held, b)

	ctx := event.C()
	if p.Handler().HandleBlockBreak(ctx, pos, &drops); ctx.Cancelled() {
		p.resendBlocks(pos, w)
		return
	}
	held, left := p.HeldItems()

	p.SwingArm()
	w.SetBlock(pos, nil, nil)
	w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})

	if breakable, ok := b.(block.Breakable); ok && !p.GameMode().CreativeInventory() {
		info := breakable.BreakInfo()
		if diff := info.XPDrops[1] - info.XPDrops[0]; diff > 0 {
			amount := rand.Intn(diff) + info.XPDrops[0]
			for _, orb := range entity.NewExperienceOrbs(pos.Vec3Centre(), amount) {
				orb.SetVelocity(mgl64.Vec3{(rand.Float64()*0.2 - 0.1) * 2, rand.Float64() * 0.4, (rand.Float64()*0.2 - 0.1) * 2})
				w.AddEntity(orb)
			}
		}
	}
	for _, drop := range drops {
		ent := entity.NewItem(drop, pos.Vec3Centre())
		ent.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
		w.AddEntity(ent)
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
	if container, ok := b.(block.Container); ok {
		// If the block is a container, it should drop its inventory contents regardless whether the
		// player is in creative mode or not.
		drops = container.Inventory().Items()
		if breakable, ok := b.(block.Breakable); ok && !p.GameMode().CreativeInventory() {
			if breakable.BreakInfo().Harvestable(t) {
				drops = breakable.BreakInfo().Drops(t, held.Enchantments())
			}
		}
		container.Inventory().Clear()
	} else if breakable, ok := b.(block.Breakable); ok && !p.GameMode().CreativeInventory() {
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

	b := p.World().Block(pos)

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

	ctx := event.C()
	if p.Handler().HandleBlockPick(ctx, pos, b); ctx.Cancelled() {
		return
	}
	_, offhand := p.HeldItems()

	if found {
		if slot < 9 {
			_ = p.session().SetHeldSlot(slot)
			return
		}
		_ = p.Inventory().Swap(slot, int(p.heldSlot.Load()))
		return
	}

	firstEmpty, emptyFound := p.Inventory().FirstEmpty()
	if !emptyFound {
		p.SetHeldItems(pickedItem, offhand)
		return
	}
	if firstEmpty < 8 {
		_ = p.session().SetHeldSlot(firstEmpty)
		_ = p.Inventory().SetItem(firstEmpty, pickedItem)
		return
	}
	_ = p.Inventory().Swap(firstEmpty, int(p.heldSlot.Load()))
	p.SetHeldItems(pickedItem, offhand)
}

// Teleport teleports the player to a target position in the world. Unlike Move, it immediately changes the
// position of the player, rather than showing an animation.
func (p *Player) Teleport(pos mgl64.Vec3) {
	ctx := event.C()
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
	p.pos.Store(pos)
	p.vel.Store(mgl64.Vec3{})
}

// Move moves the player from one position to another in the world, by adding the delta passed to the current
// position of the player.
// Move also rotates the player, adding deltaYaw and deltaPitch to the respective values.
func (p *Player) Move(deltaPos mgl64.Vec3, deltaYaw, deltaPitch float64) {
	if p.Dead() || (deltaPos.ApproxEqual(mgl64.Vec3{}) && mgl64.FloatEqual(deltaYaw, 0) && mgl64.FloatEqual(deltaPitch, 0)) {
		return
	}
	if p.immobile.Load() {
		if mgl64.FloatEqual(deltaYaw, 0) && mgl64.FloatEqual(deltaPitch, 0) {
			// If only the position was changed, don't continue with the movement when immobile.
			return
		}
		// Still update rotation if it was changed.
		deltaPos = mgl64.Vec3{}
	}
	var (
		w                     = p.World()
		pos                   = p.Position()
		yaw, pitch            = p.Rotation()
		res, resYaw, resPitch = pos.Add(deltaPos), yaw + deltaYaw, pitch + deltaPitch
	)
	ctx := event.C()
	if p.Handler().HandleMove(ctx, res, resYaw, resPitch); ctx.Cancelled() {
		if p.session() != session.Nop && pos.ApproxEqual(p.Position()) {
			// The position of the player was changed and the event cancelled. This means we still need to notify the
			// player of this movement change.
			p.teleport(pos)
		}
		return
	}
	for _, v := range p.viewers() {
		v.ViewEntityMovement(p, res, resYaw, resPitch, p.OnGround())
	}

	p.pos.Store(res)
	p.yaw.Store(resYaw)
	p.pitch.Store(resPitch)

	_, submergedBefore := w.Liquid(cube.PosFromVec3(pos.Add(mgl64.Vec3{0, p.EyeHeight()})))
	_, submergedAfter := w.Liquid(cube.PosFromVec3(res.Add(mgl64.Vec3{0, p.EyeHeight()})))
	if submergedBefore != submergedAfter {
		// Player wasn't either breathing before and no longer isn't, or wasn't breathing before and now is,
		// so send the updated metadata.
		p.session().ViewEntityState(p)
	}

	p.checkBlockCollisions(w)
	p.onGround.Store(p.checkOnGround(w))

	p.updateFallState(deltaPos[1])

	// The vertical axis isn't relevant for calculation of exhaustion points.
	deltaPos[1] = 0
	if p.Swimming() {
		p.Exhaust(0.01 * deltaPos.Len())
	} else if p.Sprinting() {
		p.Exhaust(0.1 * deltaPos.Len())
	}
}

// Facing returns the horizontal direction that the player is facing.
func (p *Player) Facing() cube.Direction {
	return entity.Facing(p)
}

// World returns the world that the player is currently in.
func (p *Player) World() *world.World {
	w, _ := world.OfEntity(p)
	return w
}

// Position returns the current position of the player. It may be changed as the player moves or is moved
// around the world.
func (p *Player) Position() mgl64.Vec3 {
	return p.pos.Load()
}

// Velocity returns the players current velocity. If there is an attached session, this will be empty.
func (p *Player) Velocity() mgl64.Vec3 {
	return p.vel.Load()
}

// SetVelocity updates the player's velocity. If there is an attached session, this will just send
// the velocity to the player session for the player to update.
func (p *Player) SetVelocity(velocity mgl64.Vec3) {
	if p.session() == session.Nop {
		p.vel.Store(velocity)
		return
	}
	for _, v := range p.viewers() {
		v.ViewEntityVelocity(p, velocity)
	}
}

// Rotation returns the yaw and pitch of the player in degrees. Yaw is horizontal rotation (rotation around the
// vertical axis, 0 when facing forward), pitch is vertical rotation (rotation around the horizontal axis, also 0
// when facing forward).
func (p *Player) Rotation() (float64, float64) {
	return p.yaw.Load(), p.pitch.Load()
}

// Collect makes the player collect the item stack passed, adding it to the inventory. The amount of items that could
// be added is returned.
func (p *Player) Collect(s item.Stack) int {
	if p.Dead() {
		return 0
	}
	ctx := event.C()
	if p.Handler().HandleItemPickup(ctx, s); ctx.Cancelled() {
		return 0
	}
	n, _ := p.Inventory().AddItem(s)
	return n
}

// Experience returns the amount of experience the player has.
func (p *Player) Experience() int {
	return p.experience.Experience()
}

// AddExperience adds experience to the player.
func (p *Player) AddExperience(amount int) int {
	ctx := event.C()
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
	p.session().SendExperience(p.experience)
	return amount
}

// RemoveExperience removes experience from the player.
func (p *Player) RemoveExperience(amount int) {
	p.experience.Remove(amount)
	p.session().SendExperience(p.experience)
}

// ExperienceLevel returns the experience level of the player.
func (p *Player) ExperienceLevel() int {
	return p.experience.Level()
}

// SetExperienceLevel sets the experience level of the player. The level must have a value between 0 and 2,147,483,647,
// otherwise the method panics.
func (p *Player) SetExperienceLevel(level int) {
	p.experience.SetLevel(level)
	p.session().SendExperience(p.experience)
}

// ExperienceProgress returns the experience progress of the player.
func (p *Player) ExperienceProgress() float64 {
	return p.experience.Progress()
}

// SetExperienceProgress sets the experience progress of the player. The progress must have a value between 0.0 and 1.0, otherwise
// the method panics.
func (p *Player) SetExperienceProgress(progress float64) {
	p.experience.SetProgress(progress)
	p.session().SendExperience(p.experience)
}

// CollectExperience makes the player collect the experience points passed, adding it to the experience manager. A bool
// is returned indicating whether the player was able to collect the experience or not, due to the 100ms delay between
// experience collection or if the player was dead or in a game mode that doesn't allow collection.
func (p *Player) CollectExperience(value int) bool {
	if p.Dead() || !p.GameMode().AllowsInteraction() {
		return false
	}
	if time.Since(p.lastXPPickup.Load()) < time.Millisecond*100 {
		return false
	}
	p.lastXPPickup.Store(time.Now())
	return p.AddExperience(value) > 0
}

// Drop makes the player drop the item.Stack passed as an entity.Item, so that it may be picked up from the
// ground.
// The dropped item entity has a pickup delay of 2 seconds.
// The number of items that was dropped in the end is returned. It is generally the count of the stack passed
// or 0 if dropping the item.Stack was cancelled.
func (p *Player) Drop(s item.Stack) int {
	e := entity.NewItem(s, p.Position().Add(mgl64.Vec3{0, 1.4}))
	e.SetVelocity(entity.DirectionVector(p).Mul(0.4))
	e.SetPickupDelay(time.Second * 2)

	ctx := event.C()
	if p.Handler().HandleItemDrop(ctx, e); ctx.Cancelled() {
		return 0
	}
	p.World().AddEntity(e)
	return s.Count()
}

// OpenBlockContainer opens a block container, such as a chest, at the position passed. If no container was
// present at that location, OpenBlockContainer does nothing.
// OpenBlockContainer will also do nothing if the player has no session connected to it.
func (p *Player) OpenBlockContainer(pos cube.Pos) {
	if p.session() != session.Nop {
		p.session().OpenBlockContainer(pos)
	}
}

// HideEntity hides a world.Entity from the Player so that it can under no circumstance see it. Hidden entities can be
// made visible again through a call to ShowEntity.
func (p *Player) HideEntity(e world.Entity) {
	if p.session() != session.Nop && p != e {
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
func (p *Player) Tick(w *world.World, current int64) {
	if p.Dead() {
		return
	}
	if p.lastTickedWorld != w {
		p.Handler().HandleChangeWorld(p.lastTickedWorld, w)
	}
	p.lastTickedWorld = w
	if _, ok := w.Liquid(cube.PosFromVec3(p.Position())); !ok {
		p.StopSwimming()
		if _, ok := p.Armour().Helmet().Item().(item.TurtleShell); ok {
			p.AddEffect(effect.New(effect.WaterBreathing{}, 1, time.Second*10))
		}
	}

	p.checkBlockCollisions(w)
	p.onGround.Store(p.checkOnGround(w))

	p.effects.Tick(p)

	p.tickFood(w)
	p.tickAirSupply(w)
	if p.Position()[1] < float64(w.Range()[0]) && p.GameMode().AllowsTakingDamage() && current%10 == 0 {
		p.Hurt(4, damage.SourceVoid{})
	}
	if !p.AttackImmune() && p.insideOfSolid(w) {
		p.Hurt(1, damage.SourceSuffocation{})
	}

	if p.OnFireDuration() > 0 {
		p.fireTicks.Sub(1)
		if !p.GameMode().AllowsTakingDamage() || p.OnFireDuration() <= 0 || w.RainingAt(cube.PosFromVec3(p.Position())) {
			p.Extinguish()
		}
		if p.OnFireDuration()%time.Second == 0 && !p.AttackImmune() {
			p.Hurt(1, damage.SourceFireTick{})
		}
	}

	if current%4 == 0 && p.usingItem.Load() {
		held, _ := p.HeldItems()
		if _, ok := held.Item().(item.Consumable); ok {
			// Eating particles seem to happen roughly every 4 ticks.
			for _, v := range p.viewers() {
				v.ViewEntityAction(p, entity.EatAction{})
			}
		}
	}

	p.cooldownMu.Lock()
	for it, ti := range p.cooldowns {
		if time.Now().After(ti) {
			delete(p.cooldowns, it)
		}
	}
	p.cooldownMu.Unlock()

	if p.session() == session.Nop && !p.Immobile() {
		m := p.mc.TickMovement(p, p.Position(), p.Velocity(), p.yaw.Load(), p.pitch.Load())
		m.Send()

		p.vel.Store(m.Velocity())
		p.Move(m.Position().Sub(p.Position()), 0, 0)
	} else {
		p.vel.Store(mgl64.Vec3{})
	}
}

// tickAirSupply tick's the player's air supply, consuming it when underwater, and replenishing it when out of water.
func (p *Player) tickAirSupply(w *world.World) {
	if !p.canBreathe(w) {
		if r, ok := p.Armour().Helmet().Enchantment(enchantment.Respiration{}); ok && rand.Float64() <= (enchantment.Respiration{}).Chance(r.Level()) {
			// Respiration grants a chance to avoid drowning damage every tick.
			return
		}

		if ticks := p.airSupplyTicks.Dec(); ticks <= -20 {
			p.airSupplyTicks.Store(0)
			if !p.AttackImmune() {
				p.Hurt(2, damage.SourceDrowning{})
			}
		}
		p.breathing = false
		p.updateState()
	} else if max := p.maxAirSupplyTicks.Load(); !p.breathing && p.airSupplyTicks.Load() < max {
		p.airSupplyTicks.Add(5)
		if p.airSupplyTicks.Load() >= max {
			p.breathing = true
			p.airSupplyTicks.Store(max)
		}
		p.updateState()
	}
}

// tickFood ticks food related functionality, such as the depletion of the food bar and regeneration if it
// is full enough.
func (p *Player) tickFood(w *world.World) {
	p.hunger.foodTick++
	if p.hunger.foodTick == 10 && (p.hunger.canQuicklyRegenerate() || w.Difficulty().FoodRegenerates()) {
		p.hunger.foodTick = 0
		p.regenerate()
		if w.Difficulty().FoodRegenerates() {
			p.AddFood(1)
		}
	} else if p.hunger.foodTick == 80 {
		p.hunger.foodTick = 0
		if p.hunger.canRegenerate() {
			p.regenerate()
		} else if p.hunger.starving() {
			p.starve(w)
		}
	}
}

// regenerate attempts to regenerate half a heart of health, typically caused by a full food bar.
func (p *Player) regenerate() {
	if p.Health() == p.MaxHealth() {
		return
	}
	p.Heal(1, healing.SourceFood{})
	p.Exhaust(6)
}

// starve deals starvation damage to the player if the difficult allows it. In peaceful mode, no damage will
// ever be dealt. In easy mode, damage will only be dealt if the player has more than 10 health. In normal
// mode, damage will only be dealt if the player has more than 2 health and in hard mode, damage will always
// be dealt.
func (p *Player) starve(w *world.World) {
	if p.Health() > w.Difficulty().StarvationHealthLimit() {
		p.Hurt(1, damage.SourceStarvation{})
	}
}

// AirSupply returns the player's remaining air supply.
func (p *Player) AirSupply() time.Duration {
	return time.Duration(p.airSupplyTicks.Load()) * time.Second / 20
}

// SetAirSupply sets the player's remaining air supply.
func (p *Player) SetAirSupply(duration time.Duration) {
	p.airSupplyTicks.Store(duration.Milliseconds() / 50)
	p.updateState()
}

// MaxAirSupply returns the player's maximum air supply.
func (p *Player) MaxAirSupply() time.Duration {
	return time.Duration(p.maxAirSupplyTicks.Load()) * time.Second / 20
}

// SetMaxAirSupply sets the player's maximum air supply.
func (p *Player) SetMaxAirSupply(duration time.Duration) {
	p.maxAirSupplyTicks.Store(duration.Milliseconds() / 50)
	p.updateState()
}

// canBreathe returns true if the player can currently breathe.
func (p *Player) canBreathe(w *world.World) bool {
	_, waterBreathing := p.effects.Effect(effect.WaterBreathing{})
	_, conduitPower := p.effects.Effect(effect.ConduitPower{})
	return waterBreathing || conduitPower || !p.insideOfWater(w)
}

// breathingDistanceBelowEyes is the lowest distance the player can be in water and still be able to breathe based on
// the player's eye height.
const breathingDistanceBelowEyes = 0.11111111

// insideOfWater returns true if the player is currently underwater.
func (p *Player) insideOfWater(w *world.World) bool {
	pos := cube.PosFromVec3(entity.EyePosition(p))
	if l, ok := w.Liquid(pos); ok {
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
func (p *Player) insideOfSolid(w *world.World) bool {
	pos := cube.PosFromVec3(entity.EyePosition(p))
	b, box := w.Block(pos), p.BBox().Translate(p.Position())

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
	for _, blockBox := range b.Model().BBox(pos, w) {
		if blockBox.Translate(pos.Vec3()).IntersectsWith(box) {
			return true
		}
	}
	return false
}

// checkCollisions checks the player's block collisions.
func (p *Player) checkBlockCollisions(w *world.World) {
	box := p.BBox().Translate(p.Position()).Grow(-0.0001)
	min, max := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())

	for y := min[1]; y <= max[1]; y++ {
		for x := min[0]; x <= max[0]; x++ {
			for z := min[2]; z <= max[2]; z++ {
				blockPos := cube.Pos{x, y, z}
				b := w.Block(blockPos)
				if collide, ok := b.(block.EntityInsider); ok {
					collide.EntityInside(blockPos, w, p)
					if _, liquid := b.(world.Liquid); liquid {
						continue
					}
				}

				if l, ok := w.Liquid(blockPos); ok {
					if collide, ok := l.(block.EntityInsider); ok {
						collide.EntityInside(blockPos, w, p)
					}
				}
			}
		}
	}
}

// checkOnGround checks if the player is currently considered to be on the ground.
func (p *Player) checkOnGround(w *world.World) bool {
	box := p.BBox().Translate(p.Position())

	b := box.Grow(1)

	min, max := cube.PosFromVec3(b.Min()), cube.PosFromVec3(b.Max())
	for x := min[0]; x <= max[0]; x++ {
		for z := min[2]; z <= max[2]; z++ {
			for y := min[1]; y < max[1]; y++ {
				pos := cube.Pos{x, y, z}
				boxList := w.Block(pos).Model().BBox(pos, w)
				for _, bb := range boxList {
					if bb.GrowVec3(mgl64.Vec3{0, 0.05}).Translate(pos.Vec3()).IntersectsWith(box) {
						return true
					}
				}
			}
		}
	}
	return false
}

// BBox returns the axis aligned bounding box of the player.
func (p *Player) BBox() cube.BBox {
	s := p.Scale()
	switch {
	// TODO: Shrink BBox for sneaking once implemented in Bedrock Edition. This is already a thing in Java Edition.
	case p.Swimming():
		return cube.Box(-0.3*s, 0, -0.3*s, 0.3*s, 0.6*s, 0.3*s)
	default:
		return cube.Box(-0.3*s, 0, -0.3*s, 0.3*s, 1.8*s, 0.3*s)
	}
}

// Scale returns the scale modifier of the Player. The default value for a normal scale is 1. A scale of 0
// will make the Player completely invisible.
func (p *Player) Scale() float64 {
	return p.scale.Load()
}

// SetScale changes the scale modifier of the Player. The default value for a normal scale is 1. A scale of 0
// will make the Player completely invisible.
func (p *Player) SetScale(s float64) {
	p.scale.Store(s)
	p.updateState()
}

// OnGround checks if the player is considered to be on the ground.
func (p *Player) OnGround() bool {
	if p.session() == session.Nop {
		return p.mc.OnGround()
	}
	return p.onGround.Load()
}

// EyeHeight returns the eye height of the player: 1.62, or 0.52 if the player is swimming.
func (p *Player) EyeHeight() float64 {
	if p.swimming.Load() {
		return 0.52
	}
	return 1.62
}

// PlaySound plays a world.Sound that only this Player can hear. Unlike World.PlaySound, it is not broadcast
// to players around it.
func (p *Player) PlaySound(sound world.Sound) {
	p.session().PlaySound(sound)
}

// ShowParticle shows a particle that only this Player can see. Unlike World.AddParticle, it is not broadcast
// to players around it.
func (p *Player) ShowParticle(pos mgl64.Vec3, particle world.Particle) {
	p.session().ViewParticle(pos, particle)
}

// EditSign edits the sign at the cube.Pos passed and writes the text passed to a sign at that position. If no sign is
// present or if the Player cannot edit it, an error is returned
func (p *Player) EditSign(pos cube.Pos, text string) error {
	w := p.World()
	sign, ok := w.Block(pos).(block.Sign)
	if !ok {
		return fmt.Errorf("edit sign: no sign at position %v", pos)
	}
	if !sign.EditableBy(p) {
		return fmt.Errorf("edit sign: sign text was already finalized")
	}

	ctx := event.C()
	if p.Handler().HandleSignEdit(ctx, sign.Text, text); ctx.Cancelled() {
		return nil
	}
	sign.Text = text
	w.SetBlock(pos, sign, nil)
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
	_, breathing := p.Effect(effect.WaterBreathing{})
	_, conduitPower := p.Effect(effect.ConduitPower{})
	_, submerged := p.World().Liquid(cube.PosFromVec3(entity.EyePosition(p)))
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
	ctx := event.C()
	if p.Handler().HandlePunchAir(ctx); ctx.Cancelled() {
		return
	}
	p.SwingArm()
	p.World().PlaySound(p.Position(), sound.Attack{})
}

// EncodeEntity ...
func (p *Player) EncodeEntity() string {
	return "minecraft:player"
}

// damageItem damages the item stack passed with the damage passed and returns the new stack. If the item
// broke, a breaking sound is played.
// If the player is not survival, the original stack is returned.
func (p *Player) damageItem(s item.Stack, d int) item.Stack {
	if p.GameMode().CreativeInventory() || d == 0 {
		return s
	}
	ctx := event.C()
	if p.Handler().HandleItemDamage(ctx, s, d); ctx.Cancelled() {
		return s
	}
	if e, ok := s.Enchantment(enchantment.Unbreaking{}); ok {
		d = (enchantment.Unbreaking{}).Reduce(s.Item(), e.Level(), d)
	}
	if s = s.Damage(d); s.Empty() {
		p.World().PlaySound(p.Position(), sound.ItemBreak{})
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
		p.dropContents()
	}
}

// canReach checks if a player can reach a position with its current range. The range depends on if the player
// is either survival or creative mode.
func (p *Player) canReach(pos mgl64.Vec3) bool {
	const (
		creativeRange = 14.0
		survivalRange = 8.0
	)
	if !p.GameMode().AllowsInteraction() {
		return false
	}
	eyes := entity.EyePosition(p)

	if p.GameMode().CreativeInventory() {
		return eyes.Sub(pos).Len() <= creativeRange && !p.Dead()
	}
	return eyes.Sub(pos).Len() <= survivalRange && !p.Dead()
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
		p.Respawn()
	}
	p.h.Swap(NopHandler{}).HandleQuit()

	if s := p.s.Swap(nil); s != nil {
		s.Disconnect(msg)
		s.CloseConnection()
		return
	}
	// Only remove the player from the world if it's not attached to a session. If it is attached to a session, the
	// session will remove the player once ready.
	p.World().RemoveEntity(p)
}

// load reads the player data from the provider. It uses the default values if the provider
// returns false.
func (p *Player) load(data Data) {
	p.yaw.Store(data.Yaw)
	p.pitch.Store(data.Pitch)

	p.health.SetMaxHealth(data.MaxHealth)
	p.health.AddHealth(data.Health - p.Health())
	p.session().SendHealth(p.health)

	p.hunger.SetFood(data.Hunger)
	p.hunger.foodTick = data.FoodTick
	p.hunger.exhaustionLevel, p.hunger.saturationLevel = data.ExhaustionLevel, data.SaturationLevel
	p.sendFood()

	p.airSupplyTicks.Store(data.AirSupply)
	p.maxAirSupplyTicks.Store(data.MaxAirSupply)

	p.experience.Add(data.Experience)
	p.session().SendExperience(p.experience)

	p.gameMode.Store(data.GameMode)
	for _, potion := range data.Effects {
		p.AddEffect(potion)
	}
	p.fireTicks.Store(data.FireTicks)
	p.fallDistance.Store(data.FallDistance)

	p.loadInventory(data.Inventory)
}

// loadInventory loads all the data associated with the player inventory.
func (p *Player) loadInventory(data InventoryData) {
	for slot, stack := range data.Items {
		_ = p.Inventory().SetItem(slot, stack)
	}
	_ = p.offHand.SetItem(0, data.OffHand)
	p.Armour().Set(data.Helmet, data.Chestplate, data.Leggings, data.Boots)
}

// Data returns the player data that needs to be saved. This is used when the player
// gets disconnected and the player provider needs to save the data.
func (p *Player) Data() Data {
	yaw, pitch := p.Rotation()
	offHand, _ := p.offHand.Item(0)

	p.hunger.mu.RLock()
	defer p.hunger.mu.RUnlock()

	return Data{
		UUID:            p.UUID(),
		Username:        p.Name(),
		Position:        p.Position(),
		Velocity:        mgl64.Vec3{},
		Yaw:             yaw,
		Pitch:           pitch,
		Health:          p.Health(),
		MaxHealth:       p.MaxHealth(),
		Hunger:          p.hunger.foodLevel,
		Experience:      p.Experience(),
		FoodTick:        p.hunger.foodTick,
		AirSupply:       p.airSupplyTicks.Load(),
		MaxAirSupply:    p.maxAirSupplyTicks.Load(),
		ExhaustionLevel: p.hunger.exhaustionLevel,
		SaturationLevel: p.hunger.saturationLevel,
		GameMode:        p.GameMode(),
		Inventory: InventoryData{
			Items:        p.Inventory().Slots(),
			Boots:        p.armour.Boots(),
			Leggings:     p.armour.Leggings(),
			Chestplate:   p.armour.Chestplate(),
			Helmet:       p.armour.Helmet(),
			OffHand:      offHand,
			MainHandSlot: p.heldSlot.Load(),
		},
		Effects:      p.Effects(),
		FireTicks:    p.fireTicks.Load(),
		FallDistance: p.fallDistance.Load(),
		Dimension:    p.World().Dimension().EncodeDimension(),
	}
}

// session returns the network session of the player. If it has one, it is returned. If not, a no-op session
// is returned.
func (p *Player) session() *session.Session {
	if s := p.s.Load(); s != nil {
		return s
	}
	return session.Nop
}

// useContext returns an item.UseContext initialised for a Player.
func (p *Player) useContext() *item.UseContext {
	call := func(ctx *event.Context, slot int, it item.Stack, f func(ctx *event.Context, slot int, it item.Stack)) error {
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
			src, dst, srcInv, dstInv := int(p.heldSlot.Load()), i, p.inv, p.armour.Inventory()
			srcIt, _ := srcInv.Item(src)
			dstIt, _ := dstInv.Item(dst)

			ctx := event.C()
			_ = call(ctx, src, srcIt, srcInv.Handler().HandleTake)
			_ = call(ctx, src, dstIt, srcInv.Handler().HandlePlace)
			_ = call(ctx, dst, dstIt, dstInv.Handler().HandleTake)
			if err := call(ctx, dst, srcIt, dstInv.Handler().HandlePlace); err == nil {
				_ = srcInv.SetItem(src, dstIt)
				_ = dstInv.SetItem(dst, srcIt)
			}
		},
		FirstFunc: func(comparable func(item.Stack) bool) (item.Stack, bool) {
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
	return p.h.Load()
}

// broadcastItems broadcasts the items held to viewers.
func (p *Player) broadcastItems(int, item.Stack) {
	for _, viewer := range p.viewers() {
		viewer.ViewEntityItems(p)
	}
}

// broadcastArmour broadcasts the armour equipped to viewers.
func (p *Player) broadcastArmour(int, item.Stack) {
	for _, viewer := range p.viewers() {
		viewer.ViewEntityArmour(p)
	}
}

// viewers returns a list of all viewers of the Player.
func (p *Player) viewers() []world.Viewer {
	viewers := p.World().Viewers(p.Position())
	var s world.Viewer = p.session()

	if sliceutil.Index(viewers, s) == -1 {
		return append(viewers, s)
	}
	return viewers
}

// resendBlocks resends blocks in a world.World at the cube.Pos passed and the block next to it at the cube.Face passed.
func (p *Player) resendBlocks(pos cube.Pos, w *world.World, faces ...cube.Face) {
	if p.session() == session.Nop {
		return
	}
	p.resendBlock(pos, w)
	for _, f := range faces {
		p.resendBlock(pos.Side(f), w)
	}
}

// resendBlock resends the block at a cube.Pos in the world.World passed.
func (p *Player) resendBlock(pos cube.Pos, w *world.World) {
	b := w.Block(pos)
	p.session().ViewBlockUpdate(pos, b, 0)
	if _, ok := b.(world.Liquid); !ok {
		if liq, ok := w.Liquid(pos); ok {
			p.session().ViewBlockUpdate(pos, liq, 1)
		}
	}
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}
