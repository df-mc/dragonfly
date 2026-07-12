package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

func TestJukeboxActivateSendsTranslatedPopup(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	u := &jukeboxTestUser{mainHand: item.NewStack(item.MusicDisc{DiscType: sound.DiscCat()}, 1)}
	ctx := &item.UseContext{}
	w.Do(func(tx *world.Tx) {
		Jukebox{}.Activate(cube.Pos{}, cube.FaceUp, tx, u, ctx)
	})

	if got, want := u.translation, chat.MessageNowPlaying; got != want {
		t.Errorf("translation: got %#v, want %#v", got, want)
	}
	if got, want := len(u.args), 1; got != want {
		t.Fatalf("argument count: got %v, want %v", got, want)
	}
	if got, want := u.args[0], any("C418 - cat"); got != want {
		t.Errorf("argument: got %q, want %q", got, want)
	}
	if got, want := ctx.CountSub, 1; got != want {
		t.Errorf("subtracted count: got %v, want %v", got, want)
	}
}

type jukeboxTestUser struct {
	mainHand    item.Stack
	translation chat.Translation
	args        []any
}

func (*jukeboxTestUser) Close() error                          { return nil }
func (*jukeboxTestUser) H() *world.EntityHandle                { return nil }
func (*jukeboxTestUser) Position() mgl64.Vec3                  { return mgl64.Vec3{} }
func (*jukeboxTestUser) Rotation() cube.Rotation               { return cube.Rotation{} }
func (u *jukeboxTestUser) HeldItems() (item.Stack, item.Stack) { return u.mainHand, item.Stack{} }
func (u *jukeboxTestUser) SetHeldItems(mainHand, _ item.Stack) { u.mainHand = mainHand }
func (*jukeboxTestUser) UsingItem() bool                       { return false }
func (*jukeboxTestUser) ReleaseItem()                          {}
func (*jukeboxTestUser) UseItem()                              {}
func (u *jukeboxTestUser) SendJukeboxPopupt(t chat.Translation, a ...any) {
	u.translation, u.args = t, a
}
