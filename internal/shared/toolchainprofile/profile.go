// Package toolchainprofile exposes the machine-readable accepted toolchain contract.
package toolchainprofile

import (
	_ "embed"
	"encoding/json"
)

// Device records the accepted device runtime and verification budgets.
type Device struct {
	Memory                      string `json:"memory"`
	HeapBytes                   uint64 `json:"heap_bytes"`
	Scheduler                   string `json:"scheduler"`
	Panic                       string `json:"panic"`
	FrameBudgetMS               uint32 `json:"frame_budget_ms"`
	RequiredSoakSeconds         uint32 `json:"required_soak_seconds"`
	OptionalExtendedSoakMinutes uint32 `json:"optional_extended_soak_minutes"`
}

// Profile records exact versions and device invariants accepted by hardware evidence.
type Profile struct {
	Schema      int    `json:"schema"`
	Go          string `json:"go"`
	PlaydateSDK string `json:"playdate_sdk"`
	TinyGo      string `json:"tinygo"`
	LLVM        string `json:"llvm"`
	ArmGCC      string `json:"arm_gcc"`
	Device      Device `json:"device"`
}

//go:embed profile.json
var encodedProfile []byte

// Accepted returns the repository's accepted immutable profile value.
func Accepted() Profile {
	var profile Profile
	if err := json.Unmarshal(encodedProfile, &profile); err != nil {
		panic("invalid embedded toolchain profile: " + err.Error())
	}
	return profile
}
