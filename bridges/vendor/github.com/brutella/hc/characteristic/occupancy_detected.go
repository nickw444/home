// THIS FILE IS AUTO-GENERATED
package characteristic

const (
	OccupancyDetectedOccupancyNotDetected int = 0
	OccupancyDetectedOccupancyDetected    int = 1
)

const TypeOccupancyDetected = "71"

type OccupancyDetected struct {
	*Int
}

func NewOccupancyDetected() *OccupancyDetected {
	char := NewInt(TypeOccupancyDetected)
	char.Format = FormatUInt8
	char.Perms = []string{PermRead, PermEvents}

	char.SetValue(0)

	return &OccupancyDetected{char}
}
