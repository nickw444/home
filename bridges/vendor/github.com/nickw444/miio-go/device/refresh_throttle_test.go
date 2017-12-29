package device

import (
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func RefreshThrottle_Setup() (tt struct {
	refreshInterval time.Duration
	clk             *clock.Mock
	throttle        *refreshThrottle
}) {
	tt.refreshInterval = 5 * time.Second
	tt.clk = clock.NewMock()
	tt.throttle = &refreshThrottle{
		tickerFactory: func() *clock.Ticker { return tt.clk.Ticker(tt.refreshInterval) },
		ch:            make(chan struct{}, 2),
	}

	return
}

//does not close Chan until Close is called
func TestRefreshThrottle_ChanDoesNotCloseUntilClose(t *testing.T) {
	tt := RefreshThrottle_Setup()

	wg := sync.WaitGroup{}
	ch := tt.throttle.Chan()
	closed := false

	wg.Add(1)

	go func() {
		ticks := 0
		for range ch {
			ticks++
		}

		assert.True(t, ticks > 1, "Expected at least 2 ticks before channel close.")

		if !closed {
			t.Error("Expected channel to not have been closed yet.")
		}

		wg.Done()
	}()

	tt.throttle.Start()
	tt.clk.Add(tt.refreshInterval)

	closed = true
	tt.throttle.Close()
	wg.Wait()
}

// Test to ensure that stop kills all clock ticks
func TestRefreshThrottle_Stop(t *testing.T) {
	tt := RefreshThrottle_Setup()

	ch := tt.throttle.Chan()
	tt.throttle.Start()
	<-ch // Clear initial tick.
	tt.clk.Add(tt.refreshInterval)
	race(t, ch)

	tt.throttle.Stop()
	tt.clk.Add(tt.refreshInterval)
	assert.Len(t, ch, 0)

	tt.clk.Add(tt.refreshInterval)
	assert.Len(t, ch, 0)
}

// Test to ensure that start starts the throttle ticking.
func TestRefreshThrottle_Start(t *testing.T) {
	tt := RefreshThrottle_Setup()

	ch := tt.throttle.Chan()
	assert.Empty(t, ch)
	tt.throttle.Start()

	// Expect an initial event.
	assert.Len(t, ch, 1)
	<-ch

	// Expect an event after refresh interval
	tt.clk.Add(tt.refreshInterval)
	race(t, ch)

	// Expect no more ticks until refresh interval
	assert.Len(t, ch, 0)
	tt.clk.Add(tt.refreshInterval - time.Second)
	assert.Len(t, ch, 0)
	tt.clk.Add(time.Second)
	assert.Len(t, ch, 1)
}

// Test to ensure that start starts the throttle ticking after being stopped.
func TestRefreshThrottle_StartAfterStop(t *testing.T) {
	tt := RefreshThrottle_Setup()

	ch := tt.throttle.Chan()
	tt.throttle.Start()
	<-ch // Clear initial tick.
	tt.clk.Add(tt.refreshInterval)
	race(t, ch)

	tt.throttle.Stop()
	tt.clk.Add(tt.refreshInterval)
	assert.Len(t, ch, 0)

	tt.clk.Add(tt.refreshInterval)
	assert.Len(t, ch, 0)

	tt.throttle.Start()
	<-ch // Clear initial tick.
	tt.clk.Add(tt.refreshInterval)
	race(t, ch)

	tt.clk.Add(tt.refreshInterval)
	assert.Len(t, ch, 1)
}

func race(t *testing.T, ch <-chan struct{}) {
	select {
	case <-ch:
		return
	case <-time.After(10 * time.Second):
		t.Error("Timed out whilst waiting for channel.")
	}
}
