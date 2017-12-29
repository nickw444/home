package device

import (
	"github.com/benbjohnson/clock"
	"time"
)

type RefreshThrottle interface {
	Chan() <-chan struct{}
	Stop()
	Start()
	Close()
}

type TickerFactory func() *clock.Ticker

type refreshThrottle struct {
	tickerFactory TickerFactory
	ch            chan struct{}

	quitChan chan struct{}
	ticker   *clock.Ticker
}

func NewRefreshThrottle(refreshInterval time.Duration) RefreshThrottle {
	c := clock.New()
	return &refreshThrottle{
		tickerFactory: func() *clock.Ticker { return c.Ticker(refreshInterval) },
		ch:            make(chan struct{}),
	}
}

func (r *refreshThrottle) Chan() <-chan struct{} {
	return r.ch
}

func (r *refreshThrottle) Start() {
	if r.ticker == nil {
		r.quitChan = make(chan struct{})
		r.ticker = r.tickerFactory()

		// Request a refresh immediately.
		r.ch <- struct{}{}
		go r.refresh()
	}
}

func (r *refreshThrottle) Stop() {
	if r.ticker != nil {
		r.ticker.Stop()
		r.ticker = nil
		close(r.quitChan)
	}
}

func (r *refreshThrottle) Close() {
	r.Stop()
	close(r.ch)
}

func (r *refreshThrottle) refresh() {
	for {
		select {
		case <-r.quitChan:
			return
		default:
		}

		select {
		case <-r.quitChan:
			return
		case <-r.ticker.C:
			r.ch <- struct{}{}
		}
	}
}
