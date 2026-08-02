package playdate

// PowerStatus describes the Playdate's current external power state.
type PowerStatus uint8

const (
	PowerCharging PowerStatus = 1 << iota
	PowerUSB
	PowerScrews
)

// Has reports whether all requested power flags are set.
func (status PowerStatus) Has(requested PowerStatus) bool { return status&requested == requested }

// Accelerometer is the optional motion-sampling capability. Sampling must be
// explicitly enabled before AccelerometerXYZ is called. The runtime disables
// it automatically after LifecycleTerminate.
type Accelerometer interface {
	SetAccelerometerEnabled(bool)
	AccelerometerXYZ() (x, y, z float32)
}

// PowerMonitor is the optional battery and external-power capability.
type PowerMonitor interface {
	PowerStatus() PowerStatus
	BatteryPercentage() float32
	BatteryVoltage() float32
}

// SystemPreferences is the optional read-only system-settings capability.
type SystemPreferences interface {
	SystemVolume() float32
	ReduceFlashing() bool
	TimezoneOffsetSeconds() int32
	Uses24HourTime() bool
}
