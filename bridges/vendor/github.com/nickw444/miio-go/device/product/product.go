package product

import "fmt"

type Product uint16

const (
	PowerPlug Product = iota << 1
	Yeelight
	Unknown
)

func GetModel(modelName string) (Product, error) {
	switch modelName {
	case "chuangmi.plug.m1":
		return PowerPlug, nil
	default:
		return Unknown, fmt.Errorf("Unknown product for device type %s", modelName)
	}
}
