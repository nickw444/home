package device

import "encoding/json"

type Vendor struct {
	ID       uint32
	Name     string
	Products map[uint32]*Product
}

func (v *Vendor) UnmarshalJSON(b []byte) error {
	var res struct {
		VID      uint32     `json:"vid"`
		Name     string     `json:"name"`
		Products []*Product `json:"products"`
	}

	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}

	v.ID = res.VID
	v.Name = res.Name
	v.Products = make(map[uint32]*Product)
	for _, product := range res.Products {
		v.Products[product.ID] = product
	}

	return nil
}
