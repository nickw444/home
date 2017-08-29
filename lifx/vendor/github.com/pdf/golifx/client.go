package golifx

import (
	"sync"
	"time"

	"github.com/pdf/golifx/common"
)

// Client provides a simple interface for interacting with LIFX devices.  Client
// can not be instantiated manually or it will not function - always use
// NewClient() to obtain a Client instance.
type Client struct {
	discoveryInterval     time.Duration
	quitChan              chan struct{}
	protocol              common.Protocol
	timeout               time.Duration
	retryInterval         time.Duration
	internalRetryInterval time.Duration
	subscriptions         map[string]*common.Subscription
	sync.RWMutex
}

// GetLocations returns a slice of all locations known to the client, or
// common.ErrNotFound if no locations are currently known.
func (c *Client) GetLocations() (locations []common.Location, err error) {
	return c.protocol.GetLocations()
}

// GetLocationByID looks up a location by its `id` and returns a common.Location.
// May return a common.ErrNotFound error if the lookup times out without finding
// the location.
func (c *Client) GetLocationByID(id string) (common.Location, error) {
	location, err := c.protocol.GetLocation(id)
	if err == nil {
		return location, nil
	}

	var timeout <-chan time.Time
	if c.timeout > 0 {
		timeout = time.After(c.timeout)
	} else {
		timeout = make(<-chan time.Time)
	}

	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = sub.Close(); err != nil {
			common.Log.Warnf("Failed closing location subscription: %+v", err)
		}
	}()
	events := sub.Events()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil, common.ErrClosed
			}
			switch event := event.(type) {
			case common.EventNewLocation:
				if id == event.Location.ID() {
					return event.Location, nil
				}
			}
		case <-timeout:
			return nil, common.ErrNotFound
		}
	}
}

// GetLocationByLabel looks up a location by its `label` and returns a
// common.Location. May return a common.ErrNotFound error if the lookup times
// out without finding the location.
func (c *Client) GetLocationByLabel(label string) (common.Location, error) {
	locations, _ := c.GetLocations()
	for _, location := range locations {
		if label == location.GetLabel() {
			return location, nil
		}
	}

	var timeout <-chan time.Time
	if c.timeout > 0 {
		timeout = time.After(c.timeout)
	} else {
		timeout = make(<-chan time.Time)
	}

	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = sub.Close(); err != nil {
			common.Log.Warnf("Failed closing location subscription: %+v", err)
		}
	}()
	events := sub.Events()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil, common.ErrClosed
			}
			switch event := event.(type) {
			case common.EventNewLocation:
				if label == event.Location.GetLabel() {
					return event.Location, nil
				}
			}
		case <-timeout:
			return nil, common.ErrNotFound
		}
	}
}

// GetGroups returns a slice of all groups known to the client, or
// common.ErrNotFound if no groups are currently known.
func (c *Client) GetGroups() (groups []common.Group, err error) {
	return c.protocol.GetGroups()
}

// GetGroupByID looks up a group by its `id` and returns a common.Group.
// May return a common.ErrNotFound error if the lookup times out without finding
// the group.
func (c *Client) GetGroupByID(id string) (common.Group, error) {
	group, err := c.protocol.GetGroup(id)
	if err == nil {
		return group, nil
	}

	var timeout <-chan time.Time
	if c.timeout > 0 {
		timeout = time.After(c.timeout)
	} else {
		timeout = make(<-chan time.Time)
	}

	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = sub.Close(); err != nil {
			common.Log.Warnf("Failed closing group subscription: %+v", err)
		}
	}()
	events := sub.Events()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil, common.ErrClosed
			}
			switch event := event.(type) {
			case common.EventNewGroup:
				if id == event.Group.ID() {
					return event.Group, nil
				}
			}
		case <-timeout:
			return nil, common.ErrNotFound
		}
	}
}

// GetGroupByLabel looks up a group by its `label` and returns a common.Group.
// May return a common.ErrNotFound error if the lookup times out without finding
// the group.
func (c *Client) GetGroupByLabel(label string) (common.Group, error) {
	groups, _ := c.GetGroups()
	for _, dev := range groups {
		if label == dev.GetLabel() {
			return dev, nil
		}
	}

	var timeout <-chan time.Time
	if c.timeout > 0 {
		timeout = time.After(c.timeout)
	} else {
		timeout = make(<-chan time.Time)
	}

	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = sub.Close(); err != nil {
			common.Log.Warnf("Failed closing group subscription: %+v", err)
		}
	}()
	events := sub.Events()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil, common.ErrClosed
			}
			switch event := event.(type) {
			case common.EventNewGroup:
				if label == event.Group.GetLabel() {
					return event.Group, nil
				}
			}
		case <-timeout:
			return nil, common.ErrNotFound
		}
	}
}

// GetDevices returns a slice of all devices known to the client, or
// common.ErrNotFound if no devices are currently known.
func (c *Client) GetDevices() (devices []common.Device, err error) {
	return c.protocol.GetDevices()
}

// GetDeviceByID looks up a device by its `id` and returns a common.Device.
// May return a common.ErrNotFound error if the lookup times out without finding
// the device.
func (c *Client) GetDeviceByID(id uint64) (common.Device, error) {
	dev, err := c.protocol.GetDevice(id)
	if err == nil {
		return dev, nil
	}

	var timeout <-chan time.Time
	if c.timeout > 0 {
		timeout = time.After(c.timeout)
	} else {
		timeout = make(<-chan time.Time)
	}

	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = sub.Close(); err != nil {
			common.Log.Warnf("Failed closing device subscription: %+v", err)
		}
	}()
	events := sub.Events()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil, common.ErrClosed
			}
			switch event := event.(type) {
			case common.EventNewDevice:
				if id == event.Device.ID() {
					return event.Device, nil
				}
			}
		case <-timeout:
			return nil, common.ErrNotFound
		}
	}
}

// GetDeviceByLabel looks up a device by its `label` and returns a common.Device.
// May return a common.ErrNotFound error if the lookup times out without finding
// the device.
func (c *Client) GetDeviceByLabel(label string) (common.Device, error) {
	devices, _ := c.GetDevices()
	for _, dev := range devices {
		res, err := dev.GetLabel()
		if err == nil && res == label {
			return dev, nil
		}
	}

	var timeout <-chan time.Time
	if c.timeout > 0 {
		timeout = time.After(c.timeout)
	} else {
		timeout = make(<-chan time.Time)
	}

	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = sub.Close(); err != nil {
			common.Log.Warnf("Failed closing device subscription: %+v", err)
		}
	}()
	events := sub.Events()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil, common.ErrClosed
			}
			switch event := event.(type) {
			case common.EventNewDevice:
				l, err := event.Device.GetLabel()
				if err != nil {
					return nil, err
				}
				if l == label {
					return event.Device, nil
				}
			}
		case <-timeout:
			return nil, common.ErrNotFound
		}
	}
}

// GetLights returns a slice of all lights known to the client, or
// common.ErrNotFound if no lights are currently known.
func (c *Client) GetLights() (lights []common.Light, err error) {
	devices, err := c.GetDevices()
	if err != nil {
		return lights, err
	}

	for _, dev := range devices {
		if light, ok := dev.(common.Light); ok {
			lights = append(lights, light)
		}
	}

	if len(lights) == 0 {
		return lights, common.ErrNotFound
	}

	return lights, nil
}

// GetLightByID looks up a light by its `id` and returns a common.Light.
// May return a common.ErrNotFound error if the lookup times out without finding
// the light, or common.ErrDeviceInvalidType if the device exists but is not a
// light.
func (c *Client) GetLightByID(id uint64) (light common.Light, err error) {
	dev, err := c.GetDeviceByID(id)
	if err != nil {
		return nil, err
	}

	light, ok := dev.(common.Light)
	if !ok {
		return nil, common.ErrDeviceInvalidType
	}

	return light, nil
}

// GetLightByLabel looks up a light by its `label` and returns a common.Light.
// May return a common.ErrNotFound error if the lookup times out without finding
// the light, or common.ErrDeviceInvalidType if the device exists but is not a
// light.
func (c *Client) GetLightByLabel(label string) (common.Light, error) {
	dev, err := c.GetDeviceByLabel(label)
	if err != nil {
		return nil, err
	}

	light, ok := dev.(common.Light)
	if !ok {
		return nil, common.ErrDeviceInvalidType
	}

	return light, nil
}

// SetPower broadcasts a request to change the power state of all devices on
// the network.  A state of true requests power on, and a state of false
// requests power off.
func (c *Client) SetPower(state bool) error {
	return c.protocol.SetPower(state)
}

// SetPowerDuration broadcasts a request to change the power state of all
// devices on the network, transitioning over the specified duration.  A state
// of true requests power on, and a state of false requests power off.  Not all
// device types support transitioning, so if you wish to change the state of all
// device types, you should use SetPower instead.
func (c *Client) SetPowerDuration(state bool, duration time.Duration) error {
	return c.protocol.SetPowerDuration(state, duration)
}

// SetColor broadcasts a request to change the color of all devices on the
// network.
func (c *Client) SetColor(color common.Color, duration time.Duration) error {
	return c.protocol.SetColor(color, duration)
}

// SetDiscoveryInterval causes the client to discover devices and state every
// interval.  You should set this to a non-zero value for any long-running
// process, otherwise devices will only be discovered once.
func (c *Client) SetDiscoveryInterval(interval time.Duration) error {
	c.Lock()
	if c.discoveryInterval != 0 {
		for i := 0; i < cap(c.quitChan); i++ {
			c.quitChan <- struct{}{}
		}
	}
	c.discoveryInterval = interval
	c.Unlock()
	common.Log.Infof("Starting discovery with interval %v", interval)
	return c.discover()
}

// SetTimeout sets the time that client operations wait for results before
// returning an error.  The special value of 0 may be set to disable timeouts,
// and all operations will wait indefinitely, but this is not recommended.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout returns the currently configured timeout period for operations on
// this client
func (c *Client) GetTimeout() *time.Duration {
	return &c.timeout
}

// SetRetryInterval sets the retry interval for operations on this client.  If
// a timeout has been set, and the retry interval exceeds the timeout, the retry
// interval will be set to half the timeout
func (c *Client) SetRetryInterval(retryInterval time.Duration) {
	if c.timeout > 0 && retryInterval >= c.timeout {
		retryInterval = c.timeout / 2
	}
	c.Lock()
	c.retryInterval = retryInterval
	c.Unlock()
}

// GetRetryInterval returns the currently configured retry interval for
// operations on this client
func (c *Client) GetRetryInterval() *time.Duration {
	c.RLock()
	defer c.RUnlock()
	return &c.retryInterval
}

// NewSubscription returns a new *common.Subscription for receiving events from
// this client.
func (c *Client) NewSubscription() (*common.Subscription, error) {
	sub := common.NewSubscription(c)
	c.Lock()
	c.subscriptions[sub.ID()] = sub
	c.Unlock()
	return sub, nil
}

// CloseSubscription is a callback for handling the closing of subscriptions.
func (c *Client) CloseSubscription(sub *common.Subscription) error {
	c.RLock()
	_, ok := c.subscriptions[sub.ID()]
	c.RUnlock()
	if !ok {
		return common.ErrNotFound
	}
	c.Lock()
	delete(c.subscriptions, sub.ID())
	c.Unlock()

	return nil
}

// Close signals the termination of this client, and cleans up resources
func (c *Client) Close() error {
	for _, sub := range c.subscriptions {
		if err := sub.Close(); err != nil {
			return err
		}
	}

	c.Lock()
	defer c.Unlock()

	select {
	case <-c.quitChan:
		common.Log.Warnf(`client already closed`)
		return common.ErrClosed
	default:
		close(c.quitChan)
	}

	return c.protocol.Close()
}

// publish an event to subscribers
func (c *Client) publish(event interface{}) error {
	c.RLock()
	subs := make(map[string]*common.Subscription, len(c.subscriptions))
	for k, sub := range c.subscriptions {
		subs[k] = sub
	}
	c.RUnlock()

	for _, sub := range subs {
		if err := sub.Write(event); err != nil {
			return err
		}
	}

	return nil
}

// subscribe to protocol events and proxy to client subscriptions
func (c *Client) subscribe() error {
	sub, err := c.protocol.NewSubscription()
	if err != nil {
		return err
	}
	events := sub.Events()

	go func() {
		for {
			select {
			case <-c.quitChan:
				return
			default:
			}
			select {
			case <-c.quitChan:
				return
			case event := <-events:
				switch event.(type) {
				case common.EventNewDevice,
					common.EventNewGroup,
					common.EventNewLocation,
					common.EventExpiredDevice,
					common.EventExpiredGroup,
					common.EventExpiredLocation:
					if err = c.publish(event); err != nil {
						common.Log.Warnf("Failed publishing event on client: %v", err)
					}
				}
			}
		}
	}()

	return nil
}

func (c *Client) discover() error {
	if c.discoveryInterval == 0 {
		common.Log.Debugf("Discovery interval is zero, discovery will only be performed once")
		return c.protocol.Discover()
	}

	go func() {
		c.RLock()
		tick := time.Tick(c.discoveryInterval)
		c.RUnlock()
		for {
			select {
			case <-c.quitChan:
				common.Log.Debugf("Quitting discovery loop")
				return
			default:
			}
			select {
			case <-c.quitChan:
				common.Log.Debugf("Quitting discovery loop")
				return
			case <-tick:
				common.Log.Debugf("Performing discovery")
				_ = c.protocol.Discover()
			}
		}
	}()

	return nil
}
