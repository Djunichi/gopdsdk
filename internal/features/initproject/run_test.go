package initproject

import (
	"runtime/debug"
	"testing"
)

func TestPublishedSDKVersion(t *testing.T) {
	tests := []struct {
		info debug.BuildInfo
		want string
	}{
		{info: debug.BuildInfo{Main: debug.Module{Path: sdkModule, Version: "v0.1.0"}}, want: "v0.1.0"},
		{info: debug.BuildInfo{Main: debug.Module{Path: sdkModule, Version: "v0.3.1-0.20260801150743-f1b552b7f309"}}},
		{info: debug.BuildInfo{Main: debug.Module{Path: sdkModule, Version: "v0.1.0"}, Settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "true"}}}},
		{info: debug.BuildInfo{Main: debug.Module{Path: sdkModule, Version: "(devel)"}}},
		{info: debug.BuildInfo{Main: debug.Module{Path: "example.com/other", Version: "v0.1.0"}}},
	}
	for _, test := range tests {
		if got := publishedSDKVersion(&test.info); got != test.want {
			t.Errorf("publishedSDKVersion(%+v) = %q, want %q", test.info, got, test.want)
		}
	}
}
