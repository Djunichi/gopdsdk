package runtime

import (
	"errors"
	"testing"

	"github.com/Djunichi/gopdsdk/playdate"
)

type statusContext struct {
	testContext
	accelerometerCalls []bool
	exited             bool
}

func (context *statusContext) SetAccelerometerEnabled(enabled bool) {
	context.accelerometerCalls = append(context.accelerometerCalls, enabled)
}
func (*statusContext) AccelerometerXYZ() (float32, float32, float32) { return 1, -2, 3 }
func (*statusContext) PowerStatus() playdate.PowerStatus {
	return playdate.PowerCharging | playdate.PowerUSB
}
func (*statusContext) BatteryPercentage() float32   { return 72.5 }
func (*statusContext) BatteryVoltage() float32      { return 4.1 }
func (*statusContext) SystemVolume() float32        { return 0.6 }
func (*statusContext) ReduceFlashing() bool         { return true }
func (*statusContext) TimezoneOffsetSeconds() int32 { return 7200 }
func (*statusContext) Uses24HourTime() bool         { return true }
func (context *statusContext) ExitToLauncher()      { context.exited = true }

type statusGame struct{}

func (statusGame) Init(context playdate.Context) error {
	accelerometer, ok := context.(playdate.Accelerometer)
	if !ok {
		return errors.New("accelerometer capability missing")
	}
	if x, y, z := accelerometer.AccelerometerXYZ(); x != 0 || y != 0 || z != 0 {
		return errors.New("accelerometer sampled before enablement")
	}
	accelerometer.SetAccelerometerEnabled(true)
	if x, y, z := accelerometer.AccelerometerXYZ(); x != 1 || y != -2 || z != 3 {
		return errors.New("accelerometer values not forwarded")
	}
	monitor, ok := context.(playdate.PowerMonitor)
	if !ok || !monitor.PowerStatus().Has(playdate.PowerCharging|playdate.PowerUSB) || monitor.BatteryPercentage() != 72.5 || monitor.BatteryVoltage() != 4.1 {
		return errors.New("power values not forwarded")
	}
	preferences, ok := context.(playdate.SystemPreferences)
	if !ok || preferences.SystemVolume() != 0.6 || !preferences.ReduceFlashing() || preferences.TimezoneOffsetSeconds() != 7200 || !preferences.Uses24HourTime() {
		return errors.New("preferences not forwarded")
	}
	launcher, ok := context.(playdate.Launcher)
	if !ok {
		return errors.New("launcher capability missing")
	}
	launcher.ExitToLauncher()
	return nil
}
func (statusGame) Update(playdate.Context) (bool, error) { return false, nil }

func TestNewApplicationForwardsStatusAndCleansUpAccelerometer(t *testing.T) {
	context := &statusContext{}
	application, err := NewApplication(statusGame{}, context, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	if !context.exited {
		t.Fatal("ExitToLauncher was not forwarded")
	}
	if err := application.Handle(EventTerminate, 0); err != nil {
		t.Fatal(err)
	}
	if len(context.accelerometerCalls) != 2 || !context.accelerometerCalls[0] || context.accelerometerCalls[1] {
		t.Fatalf("accelerometer enablement calls = %v, want [true false]", context.accelerometerCalls)
	}
}
