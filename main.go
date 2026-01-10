package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/pelletier/go-toml"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	chat.Global.Subscribe(chat.StdoutSubscriber{})
	conf, err := readConfig(slog.Default())
	if err != nil {
		panic(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()

	// Initialize the safe-fall tracker.
	// NOTE: You must call trackOnMove(...) from your movement event handler
	// and call ShouldApplyFallDamage(...) from your damage/fall event handler.
	initSafeFallTracker()

	srv.Listen()
	for p := range srv.Accept() {
		_ = p
		p.SetGameMode(world.GameModeSurvival)
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}

// Below are small helper functions and usage notes for preventing small
// (<=3 block) wind-charge / fall damage. Wire these into your server event
// handlers:
// - On player movement updates, call trackOnMove(playerID, y, onGround).
// - On fall/damage events, call adjustFallDamage(playerID, reportedDamage, currentY).
//
// The implementation below is intentionally minimal and lives in main so you
// can adapt it to the exact event API your server exposes.

// safeFall stores the last known safe Y for players.
// Key type chosen as string (player unique ID/name). Use your player's unique
// identifier when calling the functions.
var safeFall = make(map[string]float64)

// initSafeFallTracker is a placeholder to show where you'd initialize any
// per-server trackers. No-op here but kept for clarity.
func initSafeFallTracker() {
	// ...existing code...
}

// trackOnMove should be invoked from your movement update handler.
// - id: unique player id (string).
// - y: current player Y position (float64).
// - onGround: whether the server considers the player on ground.
// - inWeb: whether the player is currently in/standing on a web block (treat as safe).
//
// Update the player's last safe Y when on ground OR on web.
func trackOnMove(id string, y float64, onGround bool, inWeb bool) {
	if onGround || inWeb {
		safeFall[id] = y
	}
	// If player teleported/was wind-charged, event handler should also update
	// safeFall[id] appropriately (e.g., set to destination Y if safe).
}

// enterWeb should be called the moment a player starts intersecting a web (clutch).
// Recording the current Y here ensures clutching into a web cancels fall damage.
func enterWeb(id string, y float64) {
	// Treat entering the web as a safe position where fall distance resets.
	safeFall[id] = y
}

// clearSafeFall removes tracking for the player (call on disconnect/death).
func clearSafeFall(id string) {
	delete(safeFall, id)
}

// ShouldApplyFallDamage decides the damage to apply for fall/wind-charge damage,
// taking webs into account.
// - id: unique player id.
// - reportedDamage: damage the system is about to apply.
// - currentY: player's current Y at time of damage evaluation.
// - landedInWeb: whether the player landed/ended up in a web (web cancels fall damage).
//
// Returns adjusted damage (0 to original).
func ShouldApplyFallDamage(id string, reportedDamage float64, currentY float64, landedInWeb bool) float64 {
	if landedInWeb {
		// Landing on/into a web should cancel fall damage.
		return 0
	}
	startY, ok := safeFall[id]
	if !ok {
		return reportedDamage
	}
	// Compute fall distance from last safe Y to current Y.
	fallDistance := startY - currentY
	if fallDistance < 0 {
		fallDistance = 0
	}
	if fallDistance <= 3 {
		// Small falls (<=3 blocks) should not cause damage.
		return 0
	}
	return reportedDamage
}

// adjustFallDamage kept for backward compatibility (assumes not landing in web).
func adjustFallDamage(id string, reportedDamage float64, currentY float64) float64 {
	return ShouldApplyFallDamage(id, reportedDamage, currentY, false)
}
