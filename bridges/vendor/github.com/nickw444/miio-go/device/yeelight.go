package device

type Yeelight struct {
	Device
}

func NewYeelight(device Device) *Yeelight {
	return &Yeelight{
		Device: device,
	}
}

func (p *Yeelight) SetPower(power bool) error {
	return nil
}
