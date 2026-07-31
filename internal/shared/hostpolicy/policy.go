// Package hostpolicy defines host-specific Playdate SDK artifact and tool names.
package hostpolicy

import (
	"fmt"
	"path/filepath"
)

// Policy describes one supported development host.
type Policy struct {
	GOOS                string
	LibraryExtension    string
	PDCName             string
	PDUtilName          string
	CompilerCandidates  []string
	SimulatorCandidates []string
}

// For returns the policy for a Go host operating system.
func For(goos string) (Policy, error) {
	switch goos {
	case "windows":
		return Policy{GOOS: goos, LibraryExtension: "dll", PDCName: "pdc.exe", PDUtilName: "pdutil.exe", CompilerCandidates: []string{"gcc", "clang", "cc"}, SimulatorCandidates: []string{filepath.Join("bin", "PlaydateSimulator.exe")}}, nil
	case "darwin":
		return Policy{GOOS: goos, LibraryExtension: "dylib", PDCName: "pdc", PDUtilName: "pdutil", CompilerCandidates: []string{"clang", "cc", "gcc"}, SimulatorCandidates: []string{filepath.Join("bin", "Playdate Simulator", "Contents", "MacOS", "Playdate Simulator"), filepath.Join("bin", "Playdate Simulator.app", "Contents", "MacOS", "Playdate Simulator")}}, nil
	case "linux":
		return Policy{GOOS: goos, LibraryExtension: "so", PDCName: "pdc", PDUtilName: "pdutil", CompilerCandidates: []string{"gcc", "clang", "cc"}, SimulatorCandidates: []string{filepath.Join("bin", "PlaydateSimulator")}}, nil
	default:
		return Policy{}, fmt.Errorf("unsupported host OS %q", goos)
	}
}
