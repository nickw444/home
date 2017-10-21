// THIS FILE IS AUTO-GENERATED
package characteristic

const (
	LockCurrentStateUnsecured int = 0
	LockCurrentStateSecured   int = 1
	LockCurrentStateJammed    int = 2
	LockCurrentStateUnknown   int = 3
)

const TypeLockCurrentState = "1D"

type LockCurrentState struct {
	*Int
}

func NewLockCurrentState() *LockCurrentState {
	char := NewInt(TypeLockCurrentState)
	char.Format = FormatUInt8
	char.Perms = []string{PermRead, PermEvents}

	char.SetValue(0)

	return &LockCurrentState{char}
}
