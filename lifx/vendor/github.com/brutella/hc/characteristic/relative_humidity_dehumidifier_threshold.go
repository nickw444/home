// THIS FILE IS AUTO-GENERATED
package characteristic

const TypeRelativeHumidityDehumidifierThreshold = "C9"

type RelativeHumidityDehumidifierThreshold struct {
	*Float
}

func NewRelativeHumidityDehumidifierThreshold() *RelativeHumidityDehumidifierThreshold {
	char := NewFloat(TypeRelativeHumidityDehumidifierThreshold)
	char.Format = FormatFloat
	char.Perms = []string{PermRead, PermWrite, PermEvents}
	char.SetMinValue(0)
	char.SetMaxValue(100)
	char.SetStepValue(1)
	char.SetValue(0)

	return &RelativeHumidityDehumidifierThreshold{char}
}
