package entity

import (
	"testing"
	"time"
)

func TestTravelComputerStopTravelling(t *testing.T) {
	t.Run("keeps timer after portal contact", func(t *testing.T) {
		tc := &TravelComputer{inside: true, awaitingTravel: true, start: time.Now()}
		tc.StopTravelling()
		if !tc.awaitingTravel {
			t.Fatal("StopTravelling() reset travel timer after portal contact")
		}
		if tc.inside {
			t.Fatal("StopTravelling() did not clear portal contact for the next tick")
		}
	})

	t.Run("resets timer without portal contact", func(t *testing.T) {
		tc := &TravelComputer{awaitingTravel: true, start: time.Now()}
		tc.StopTravelling()
		if tc.awaitingTravel {
			t.Fatal("StopTravelling() kept travel timer without portal contact")
		}
	})
}
