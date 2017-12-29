package miio

import (
	"sync"
	"time"

	"github.com/nickw444/miio-go/protocol"
	"github.com/nickw444/miio-go/subscription"
)

type Client struct {
	sync.RWMutex
	subscription.SubscriptionTarget

	protocol          protocol.Protocol
	discoveryInterval time.Duration
	quitChan          chan struct{}
	events            chan interface{}
}

// NewClient creates a new default Client with the protocol.
func NewClient() (*Client, error) {
	p, err := protocol.NewProtocol(nil)
	if err != nil {
		return nil, err
	}

	return NewClientWithProtocol(p)
}

func NewClientWithProtocol(protocol protocol.Protocol) (*Client, error) {
	c := &Client{
		SubscriptionTarget: subscription.NewTarget(),
		protocol:           protocol,
		quitChan:           make(chan struct{}),
	}

	return c, c.init()
}

func (c *Client) init() error {
	if err := c.subscribe(); err != nil {
		return err
	}
	return c.discover()
}

func (c *Client) SetDiscoveryInterval(interval time.Duration) error {
	// TODO Clean up existing discovery goroutines.
	c.discoveryInterval = interval
	c.protocol.SetExpiryTime(interval * 3)
	return c.discover()
}

func (c *Client) discover() error {
	if c.discoveryInterval == 0 {
		//common.Log.Debugf("Discovery interval is zero, discovery will only be performed once")
		return c.protocol.Discover()
	}

	go func() {
		c.RLock()
		tick := time.Tick(c.discoveryInterval)
		c.RUnlock()
		for {
			select {
			case <-c.quitChan:
				//common.Log.Debugf("Quitting discoveryfile loop")
				return
			default:
			}
			select {
			case <-c.quitChan:
				//common.Log.Debugf("Quitting discovery loop")
				return
			case <-tick:
				//common.Log.Debugf("Performing discovery")
				_ = c.protocol.Discover()
			}
		}
	}()

	return nil
}

// Proxy events from protocol level
func (c *Client) subscribe() error {
	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return err
	}

	go func() {
		for event := range sub.Events() {
			c.Publish(event)
		}
	}()
	return nil
}
