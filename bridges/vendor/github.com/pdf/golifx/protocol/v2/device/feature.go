//go:generate enumer -type=Feature
package device

// Feature flags for devices
type Feature uint32

const (
	FeatureLight Feature = iota << 1
	FeatureColor
	FeatureInfrared
	FeatureMultizone
)
