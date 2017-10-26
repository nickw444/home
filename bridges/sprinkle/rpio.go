package main

// A wrapper for RPIO to allow injection and faking.

import (
	"github.com/Sirupsen/logrus"
	"github.com/stianeikeland/go-rpio"
)

type Direction rpio.Direction

const (
	Input Direction = iota
	Output
)

type Pin rpio.Pin

type RPIO interface {
	Open() error
	Close() error
	PinMode(pin int, direction Direction)
	SetPinHigh(pin int)
	SetPinLow(pin int)
}

type HardwareRPIO struct{}

func NewHardwareRPIO() *HardwareRPIO {
	return &HardwareRPIO{}
}

func (b *HardwareRPIO) Open() error {
	return rpio.Open()
}

func (b *HardwareRPIO) Close() error {
	return rpio.Close()
}

func (b *HardwareRPIO) PinMode(pin int, direction Direction) {
	var rpioDir rpio.Direction
	if direction == Input {
		rpioDir = rpio.Input
	} else if direction == Output {
		rpioDir = rpio.Output
	}

	rpio.PinMode(rpio.Pin(pin), rpioDir)
}

func (b *HardwareRPIO) SetPinHigh(pin int) {
	rpio.WritePin(rpio.Pin(pin), rpio.High)
}

func (b *HardwareRPIO) SetPinLow(pin int) {
	rpio.WritePin(rpio.Pin(pin), rpio.Low)
}

type FakeRPIO struct {
	log *logrus.Entry
}

func NewFakeRPIO() *FakeRPIO {
	return &FakeRPIO{
		log: log.WithField("logger", "FakeRPIO"),
	}
}

func (f *FakeRPIO) Open() error {
	f.log.Println("GPIO Opened")
	return nil
}

func (f *FakeRPIO) Close() error {
	f.log.Println("GPIO Closed")
	return nil
}

func (f *FakeRPIO) PinMode(pin int, direction Direction) {
	var rpioDir string
	if direction == Input {
		rpioDir = "INPUT"
	} else if direction == Output {
		rpioDir = "OUTPUT"
	}
	f.log.Printf("PinMode for pin %d set to %s", pin, rpioDir)
}

func (f *FakeRPIO) SetPinHigh(pin int) {
	f.log.Printf("Set pin %d HIGH", pin)
}

func (f *FakeRPIO) SetPinLow(pin int) {
	f.log.Printf("Set pin %d LOW", pin)
}
