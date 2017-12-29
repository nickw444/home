package device

import "encoding/json"

type Product struct {
	ID       uint32
	VendorID uint32
	Name     string
	Features Feature
}

func (p *Product) Supports(feature Feature) bool {
	return p.Features&feature == feature
}

func (p *Product) UnmarshalJSON(b []byte) error {
	var res struct {
		PID      uint32          `json:"pid"`
		Name     string          `json:"name"`
		Features map[string]bool `json:"features"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}
	p.ID = res.PID
	p.Name = res.Name

	// All current devices are lights, add conditional when this is not the case
	p.Features |= FeatureLight

	if res.Features[`color`] {
		p.Features |= FeatureColor
	}
	if res.Features[`infrared`] {
		p.Features |= FeatureInfrared
	}
	if res.Features[`multizone`] {
		p.Features |= FeatureMultizone
	}

	return nil
}
