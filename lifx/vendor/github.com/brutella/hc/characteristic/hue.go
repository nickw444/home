// THIS FILE IS AUTO-GENERATED
package characteristic

const TypeHue = "13"

type Hue struct {
	*Float
}

func NewHue() *Hue {
	char := NewFloat(TypeHue)
	char.Format = FormatFloat
	char.Perms = []string{PermRead, PermWrite, PermEvents}
	char.SetMinValue(0)
	char.SetMaxValue(360)
	char.SetStepValue(1)
	char.SetValue(0)
	char.Unit = UnitArcDegrees

	return &Hue{char}
}
