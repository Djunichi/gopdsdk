package playdate

import "testing"

func TestPowerStatusHas(t *testing.T) {
	status := PowerCharging | PowerUSB
	if !status.Has(PowerCharging) || !status.Has(PowerCharging|PowerUSB) || status.Has(PowerScrews) {
		t.Fatalf("unexpected power flag matching for %08b", status)
	}
}
