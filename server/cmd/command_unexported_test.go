package cmd_test

import (
	"sync/atomic"
	"testing"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// testSource is a minimal cmd.Source (Target + SendCommandOutput) for executing
// a Runnable in a unit test.
type testSource struct{}

func (testSource) Position() mgl64.Vec3          { return mgl64.Vec3{} }
func (testSource) SendCommandOutput(*cmd.Output) {}

// captureRunnable carries per-registration state in an UNEXPORTED field (not a
// cmd: parameter). It records into seen the value it observes when Run executes.
type captureRunnable struct {
	secret string
	seen   *atomic.Value
}

func (c captureRunnable) Run(cmd.Source, *cmd.Output, *world.Tx) {
	c.seen.Store(c.secret)
}

// TestCommand_PreservesUnexportedRunnableState locks a behaviour downstream code
// relies on: a Runnable may keep non-parameter state in unexported fields, and
// that state must survive into Run. New stores the runnable value as-is, and
// Execute copies it into the executed instance (overwriting only cmd:-tagged
// exported fields) rather than zero-initialising it. If this regresses (e.g.
// Execute starts from a zero value), callers that register many commands sharing
// one Runnable type but differing by an unexported field would all misbehave.
func TestCommand_PreservesUnexportedRunnableState(t *testing.T) {
	var seen atomic.Value
	c := cmd.New("probe", "probe", nil, captureRunnable{secret: "kept", seen: &seen})

	// Empty args: the runnable has no exported parameters, so no parsing occurs
	// and tx is unused.
	c.Execute("", testSource{}, nil)

	got, _ := seen.Load().(string)
	if got != "kept" {
		t.Fatalf("unexported runnable field not preserved into Run: got %q, want %q", got, "kept")
	}
}
