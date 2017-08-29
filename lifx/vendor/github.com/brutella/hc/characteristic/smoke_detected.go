// THIS FILE IS AUTO-GENERATED
package characteristic

const (
	SmokeDetectedSmokeNotDetected int = 0
	SmokeDetectedSmokeDetected    int = 1
)

const TypeSmokeDetected = "76"

type SmokeDetected struct {
	*Int
}

func NewSmokeDetected() *SmokeDetected {
	char := NewInt(TypeSmokeDetected)
	char.Format = FormatUInt8
	char.Perms = []string{PermRead, PermEvents}

	char.SetValue(0)

	return &SmokeDetected{char}
}
