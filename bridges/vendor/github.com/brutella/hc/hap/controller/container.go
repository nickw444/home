package controller

import (
	"bytes"
	"encoding/json"
	"github.com/brutella/hc/accessory"
	"io"
)

// ContainerController implements the AccessoriesHandler interface.
type ContainerController struct {
	container *accessory.Container
}

// NewContainerController returns a controller for the argument container.
func NewContainerController(m *accessory.Container) *ContainerController {
	return &ContainerController{container: m}
}

// HandleGetAccessories returns the container as json bytes.
func (ctr *ContainerController) HandleGetAccessories(r io.Reader) (io.Reader, error) {
	result, err := json.Marshal(ctr.container)
	return bytes.NewBuffer(result), err
}

// IdentifyAccessory calls Identify() for all accessories.
func (ctr *ContainerController) IdentifyAccessory() {
	for _, a := range ctr.container.Accessories {
		a.Identify()
	}
}
