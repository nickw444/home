package main

import (
	"fmt"
)

type BlindAction struct {
	Name  string
	Value uint8
}

var validActions = []BlindAction{
	BlindAction{Value: 127, Name: "PAIR"},
	BlindAction{Value: 252, Name: "DOWN"},
	BlindAction{Value: 253, Name: "STOP"},
	BlindAction{Value: 254, Name: "UP"},
}

func ActionFromValue(value uint8) (BlindAction, error) {
	for _, action := range validActions {
		if action.Value == value {
			return action, nil
		}
	}
	return BlindAction{}, fmt.Errorf("Invalid action value passed to ActionFromValue")
}

func ActionFromName(name string) (BlindAction, error) {
	for _, action := range validActions {
		if action.Name == name {
			return action, nil
		}
	}
	return BlindAction{}, fmt.Errorf("Invalid action name passed to ActionFromName")
}

func GetValidNames() []string {
	names := []string{}
	for _, action := range validActions {
		names = append(names, action.Name)
	}
	return names
}
