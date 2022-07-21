package session

import (
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"time"
)

// SendMessage ...
func (s *Session) SendMessage(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeRaw,
		Message:  message,
	})
}

// SendTip ...
func (s *Session) SendTip(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeTip,
		Message:  message,
	})
}

// SendAnnouncement ...
func (s *Session) SendAnnouncement(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeAnnouncement,
		Message:  message,
	})
}

// SendPopup ...
func (s *Session) SendPopup(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypePopup,
		Message:  message,
	})
}

// SendJukeboxPopup ...
func (s *Session) SendJukeboxPopup(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeJukeboxPopup,
		Message:  message,
	})
}

// SendToast ...
func (s *Session) SendToast(title, message string) {
	s.writePacket(&packet.ToastRequest{
		Title:   title,
		Message: message,
	})
}

// SendScoreboard ...
func (s *Session) SendScoreboard(sb *scoreboard.Scoreboard) {
	if s == Nop {
		return
	}
	currentName, currentLines := s.currentScoreboard.Load(), s.currentLines.Load()

	if currentName != sb.Name() {
		s.RemoveScoreboard()
		s.writePacket(&packet.SetDisplayObjective{
			DisplaySlot:   "sidebar",
			ObjectiveName: sb.Name(),
			DisplayName:   sb.Name(),
			CriteriaName:  "dummy",
		})
		s.currentScoreboard.Store(sb.Name())
		s.currentLines.Store(append([]string(nil), sb.Lines()...))
	} else {
		// Remove all current lines from the scoreboard. We can't replace them without removing them.
		pk := &packet.SetScore{ActionType: packet.ScoreboardActionRemove}
		for i := range currentLines {
			pk.Entries = append(pk.Entries, protocol.ScoreboardEntry{
				EntryID:       int64(i),
				ObjectiveName: currentName,
				Score:         int32(i),
			})
		}
		if len(pk.Entries) > 0 {
			s.writePacket(pk)
		}
	}
	pk := &packet.SetScore{ActionType: packet.ScoreboardActionModify}
	for k, line := range sb.Lines() {
		if len(line) == 0 {
			line = "§" + colours[k]
		}
		pk.Entries = append(pk.Entries, protocol.ScoreboardEntry{
			EntryID:       int64(k),
			ObjectiveName: sb.Name(),
			Score:         int32(k),
			IdentityType:  protocol.ScoreboardIdentityFakePlayer,
			DisplayName:   line,
		})
	}
	if len(pk.Entries) > 0 {
		s.writePacket(pk)
	}
}

// colours holds a list of colour codes to be filled out for empty lines in a scoreboard.
var colours = [15]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

// RemoveScoreboard ...
func (s *Session) RemoveScoreboard() {
	s.writePacket(&packet.RemoveObjective{ObjectiveName: s.currentScoreboard.Load()})
	s.currentScoreboard.Store("")
	s.currentLines.Store([]string{})
}

// SendBossBar sends a boss bar to the player with the text passed and the health percentage of the bar.
// SendBossBar removes any boss bar that might be active before sending the new one.
func (s *Session) SendBossBar(text string, colour uint8, healthPercentage float64) {
	s.RemoveBossBar()
	s.writePacket(&packet.BossEvent{
		BossEntityUniqueID: selfEntityRuntimeID,
		EventType:          packet.BossEventShow,
		BossBarTitle:       text,
		HealthPercentage:   float32(healthPercentage),
		Colour:             uint32(colour),
	})
}

// RemoveBossBar removes any boss bar currently active on the player's screen.
func (s *Session) RemoveBossBar() {
	s.writePacket(&packet.BossEvent{
		BossEntityUniqueID: selfEntityRuntimeID,
		EventType:          packet.BossEventHide,
	})
}

const tickLength = time.Second / 20

// SetTitleDurations ...
func (s *Session) SetTitleDurations(fadeInDuration, remainDuration, fadeOutDuration time.Duration) {
	s.writePacket(&packet.SetTitle{
		ActionType:      packet.TitleActionSetDurations,
		FadeInDuration:  int32(fadeInDuration / tickLength),
		RemainDuration:  int32(remainDuration / tickLength),
		FadeOutDuration: int32(fadeOutDuration / tickLength),
	})
}

// SendTitle ...
func (s *Session) SendTitle(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetTitle, Text: text})
}

// SendSubtitle ...
func (s *Session) SendSubtitle(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetSubtitle, Text: text})
}

// SendActionBarMessage ...
func (s *Session) SendActionBarMessage(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetActionBar, Text: text})
}
