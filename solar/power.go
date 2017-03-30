package main

// With inspiration from: https://github.com/brutella/hksymo

import (
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"log"
	"math"
)

const (
	TypePower      = "032B12CF-D4E8-4277-9021-188816FD00C6"
	TypeTotalPower = "032B12CA-D4E8-4277-9021-188816FD00C6"
	TypeInverter   = "14FA9D31-FC94-4F98-B00D-4AE878523748"
)

type Power struct {
	*characteristic.Int
}

func NewPower(val int) *Power {
	p := Power{characteristic.NewInt("")}
	p.Value = val
	p.Format = characteristic.FormatUInt64
	p.Perms = characteristic.PermsRead()
	return &p
}

type PowerService struct {
	*service.Service

	EfergyClient *EfergyClient
	Name         *characteristic.Name

	// Total House Load
	TotalLoad *Power

	// Input from different sources
	SolarInput   *Power
	MainsInput   *Power
	BatteryInput *Power

	// Selling back to the grid
	MainsOutput *Power
}

func NewPowerService(name string, efergyClient *EfergyClient) *PowerService {
	nameChar := characteristic.NewName()
	nameChar.SetValue(name)

	totalLoad := NewPower(0)
	totalLoad.Type = TypeTotalPower
	totalLoad.Unit = "W"
	totalLoad.Description = "Total Load"

	solarInput := NewPower(0)
	solarInput.Type = TypeTotalPower
	solarInput.Unit = "W"
	solarInput.Description = "Solar Input"

	mainsInput := NewPower(0)
	mainsInput.Type = TypeTotalPower
	mainsInput.Unit = "W"
	mainsInput.Description = "Mains Input"

	batteryInput := NewPower(0)
	batteryInput.Type = TypeTotalPower
	batteryInput.Unit = "W"
	batteryInput.Description = "Battery Input"

	mainsOutput := NewPower(0)
	mainsOutput.Type = TypeTotalPower
	mainsOutput.Unit = "W"
	mainsOutput.Description = "Mains Output"

	svc := service.New(TypeInverter)

	svc.AddCharacteristic(nameChar.Characteristic)
	svc.AddCharacteristic(totalLoad.Characteristic)
	svc.AddCharacteristic(solarInput.Characteristic)
	svc.AddCharacteristic(mainsInput.Characteristic)
	svc.AddCharacteristic(batteryInput.Characteristic)
	svc.AddCharacteristic(mainsOutput.Characteristic)

	return &PowerService{
		Service:      svc,
		EfergyClient: efergyClient,
		Name:         nameChar,
		TotalLoad:    totalLoad,
		SolarInput:   solarInput,
		MainsInput:   mainsInput,
		BatteryInput: batteryInput,
		MainsOutput:  mainsOutput,
	}
}

type OrganisedReadings struct {
	LoadACTong     Reading
	LoadBTong      Reading
	SolarInputTong Reading
}

const (
	SidACTong    = "786148"
	SidBTong     = "784242"
	SidSolarTong = "731181"
)

func NewOrganisedReadings(readings []*Reading) *OrganisedReadings {
	org := &OrganisedReadings{}

	for _, reading := range readings {
		switch reading.Sid {
		case SidACTong:
			org.LoadACTong = *reading
			break

		case SidBTong:
			org.LoadBTong = *reading
			break

		case SidSolarTong:
			org.SolarInputTong = *reading
			break
		}
	}

	return org
}

func (svc *PowerService) Update() error {
	log.Printf("Updating...")
	readings, err := svc.EfergyClient.GetLatestReadings()
	if err != nil {
		return err
	}

	org := NewOrganisedReadings(readings)

	totalLoad := org.LoadACTong.LastValue + org.LoadBTong.LastValue
	solarInput := int(math.Max(float64(org.SolarInputTong.LastValue-85), 0))

	// Mains Input = Load from A/C + Any additional load on
	mainsInput := org.LoadACTong.LastValue
	mainsOutput := 0
	if org.LoadBTong.LastValue > solarInput {
		// B phase consuming more power than solar producing. Excess is being
		// drawn from the grid.
		mainsInput += org.LoadBTong.LastValue - solarInput
	} else {
		// B phase not drawing enough. Exporting to the grid.
		mainsOutput += solarInput - org.LoadBTong.LastValue
	}

	batteryInput := 0

	svc.TotalLoad.Characteristic.UpdateValue(totalLoad)
	svc.SolarInput.Characteristic.UpdateValue(solarInput)
	svc.MainsInput.Characteristic.UpdateValue(mainsInput)
	svc.BatteryInput.Characteristic.UpdateValue(batteryInput)
	svc.MainsOutput.Characteristic.UpdateValue(mainsOutput)

	return nil
}
