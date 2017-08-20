// THIS FILE IS AUTO-GENERATED
package characteristic

const TypeProgrammableSwitchEvent = "73"

type ProgrammableSwitchEvent struct {
	*Int
}

func NewProgrammableSwitchEvent() *ProgrammableSwitchEvent {
	char := NewInt(TypeProgrammableSwitchEvent)
	char.Format = FormatUInt8
	char.Perms = []string{PermRead, PermEvents}
	char.SetMinValue(0)
	char.SetMaxValue(1)
	char.SetStepValue(1)
	char.SetValue(0)

	return &ProgrammableSwitchEvent{char}
}
