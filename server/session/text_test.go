package session

import (
	"reflect"
	"testing"

	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/text/language"
)

func TestSendJukeboxTranslation(t *testing.T) {
	s := &Session{
		packets:         make(chan packet.Packet, 1),
		closeBackground: make(chan struct{}),
	}
	s.SendJukeboxTranslation(chat.MessageNowPlaying, language.French, []any{"C418 - cat"})

	pk, ok := (<-s.packets).(*packet.Text)
	if !ok {
		t.Fatalf("expected *packet.Text, got %T", pk)
	}
	if got, want := pk.TextType, byte(packet.TextTypeJukeboxPopup); got != want {
		t.Errorf("text type: got %v, want %v", got, want)
	}
	if !pk.NeedsTranslation {
		t.Error("expected jukebox popup to require translation")
	}
	if got, want := pk.Message, "§r%record.nowPlaying"; got != want {
		t.Errorf("message: got %q, want %q", got, want)
	}
	if got, want := pk.Parameters, []string{"C418 - cat"}; !reflect.DeepEqual(got, want) {
		t.Errorf("parameters: got %q, want %q", got, want)
	}
}
