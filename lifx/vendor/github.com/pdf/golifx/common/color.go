package common

import "math"

// Color is used to represent the color and color temperature of a light.
// The color is represented as a 48-bit HSB (Hue, Saturation, Brightness) value.
// The color temperature is represented in K (Kelvin) and is used to adjust the
// warmness / coolness of a white light, which is most obvious when saturation
// is close zero.
type Color struct {
	Hue        uint16 `json:"hue"`        // range 0 to 65535
	Saturation uint16 `json:"saturation"` // range 0 to 65535
	Brightness uint16 `json:"brightness"` // range 0 to 65535
	Kelvin     uint16 `json:"kelvin"`     // range 2500° (warm) to 9000° (cool)
}

// AverageColor returns the average of the provided colors
func AverageColor(colors ...Color) (color Color) {
	var (
		x, y               float64
		hue, sat, bri, kel int
	)

	// Sum sind/cosd for hues
	for _, c := range colors {
		// Convert hue to degrees
		h := float64(c.Hue) / float64(math.MaxUint16) * 360.0

		x += math.Cos(h / 180.0 * math.Pi)
		y += math.Sin(h / 180.0 * math.Pi)
		sat += int(c.Saturation)
		bri += int(c.Brightness)
		kel += int(c.Kelvin)
	}

	// Average sind/cosd
	x /= float64(len(colors))
	y /= float64(len(colors))

	// Take atan2 of averaged hue and convert to uint16 scale
	hue = int((math.Atan2(y, x) * 180.0 / math.Pi) / 360.0 * float64(math.MaxUint16))
	sat /= len(colors)
	bri /= len(colors)
	kel /= len(colors)

	color.Hue = uint16(hue)
	color.Saturation = uint16(sat)
	color.Brightness = uint16(bri)
	color.Kelvin = uint16(kel)

	return color
}

// ColorEqual tests whether two Colors are equal
func ColorEqual(a, b Color) bool {
	return a.Hue == b.Hue &&
		a.Saturation == b.Saturation &&
		a.Brightness == b.Brightness &&
		a.Kelvin == b.Kelvin
}
