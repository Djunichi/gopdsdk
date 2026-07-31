package doctor

import "strings"

const (
	verifiedGoVersion     = "go1.26.5"
	verifiedSDKVersion    = "3.1.1"
	verifiedTinyGoVersion = "0.41.1"
	verifiedArmGCCVersion = "15.3.1"
)

func matchesVerifiedTool(name, version string) bool {
	switch name {
	case "go":
		return strings.Contains(version, "go version "+verifiedGoVersion+" ")
	case "tinygo":
		return strings.Contains(version, "tinygo version "+verifiedTinyGoVersion+" ")
	case "arm-none-eabi-gcc":
		return strings.Contains(version, ") "+verifiedArmGCCVersion+" ") || strings.HasSuffix(version, ") "+verifiedArmGCCVersion)
	default:
		return false
	}
}

func detectedToolchainSummary(report Report) string {
	var parts []string
	for _, name := range []string{"tinygo", "arm-none-eabi-gcc"} {
		if tool, ok := report.tool(name); ok && tool.Version != "" {
			parts = append(parts, tool.Version)
		}
	}
	return strings.Join(parts, "; ")
}
