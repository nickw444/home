// Copyright 2015 Peter Fern
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file

// Package golifx provides a simple Go interface to the LIFX LAN protocol.
//
// Based on the protocol documentation available at:
// http://lan.developer.lifx.com/
//
// Also included in cmd/lifx is a small CLI utility that allows interacting with
// your LIFX devices on the LAN.
//
// In various parts of this package you may find references to a Device or a
// Light.  The LIFX protocol makes room for future non-light devices by making a
// light a superset of a device, so a Light is a Device, but a Device is not
// necessarily a Light.  At this stage, LIFX only produces lights though, so
// they are the only type of device you will interact with.
package golifx

import (
	"time"

	"github.com/pdf/golifx/common"
)

const (
	// VERSION of this library
	VERSION = "0.5.1"
)

// NewClient returns a pointer to a new Client and any error that occurred
// initializing the client, using the protocol p.  It also kicks off a discovery
// run.
func NewClient(p common.Protocol) (*Client, error) {
	c := &Client{
		protocol:              p,
		subscriptions:         make(map[string]*common.Subscription),
		timeout:               common.DefaultTimeout,
		retryInterval:         common.DefaultRetryInterval,
		internalRetryInterval: 10 * time.Millisecond,
		quitChan:              make(chan struct{}, 2),
	}
	c.protocol.SetTimeout(&c.timeout)
	c.protocol.SetRetryInterval(&c.retryInterval)
	if err := c.subscribe(); err != nil {
		return nil, err
	}
	err := c.discover()
	return c, err
}

// SetLogger allows assigning a custom levelled logger that conforms to the
// common.Logger interface.  To capture logs generated during client creation,
// this should be called before creating a Client. Defaults to
// common.StubLogger, which does no logging at all.
func SetLogger(logger common.Logger) {
	common.SetLogger(logger)
}
