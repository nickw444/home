package endpoint

import (
	"github.com/brutella/hc/event"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/hap/pair"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/util"
	"io"
	"net/http"
)

// Pairing handles the /pairings endpoint.
//
// This endpoint is not session based and the same for all connections.
type Pairing struct {
	http.Handler

	controller *pair.PairingController
	emitter    event.Emitter
}

// NewPairing returns a new handler for pairing enpdoint
func NewPairing(controller *pair.PairingController, emitter event.Emitter) *Pairing {
	endpoint := Pairing{
		controller: controller,
		emitter:    emitter,
	}

	return &endpoint
}

func (endpoint *Pairing) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	log.Debug.Printf("%v POST /pairings", request.RemoteAddr)
	response.Header().Set("Content-Type", hap.HTTPContentTypePairingTLV8)

	var err error
	var in util.Container
	var out util.Container

	if in, err = util.NewTLV8ContainerFromReader(request.Body); err == nil {
		out, err = endpoint.controller.Handle(in)
	}

	if err != nil {
		log.Info.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
	} else {
		io.Copy(response, out.BytesBuffer())

		// Send events based on pairing method type
		b := in.GetByte(pair.TagPairingMethod)
		switch pair.PairMethodType(b) {
		case pair.PairingMethodDelete: // pairing removed
			endpoint.emitter.Emit(event.DeviceUnpaired{})

		case pair.PairingMethodAdd: // pairing added
			endpoint.emitter.Emit(event.DevicePaired{})

		}
	}
}
