package doctor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type fakeSystem struct {
	goos    string
	goarch  string
	home    string
	env     map[string]string
	tools   map[string]string
	homeErr error
}

func (f fakeSystem) GOOS() string                          { return f.goos }
func (f fakeSystem) GOARCH() string                        { return f.goarch }
func (f fakeSystem) HomeDir() (string, error)              { return f.home, f.homeErr }
func (f fakeSystem) ReadFile(path string) ([]byte, error)  { return os.ReadFile(path) }
func (f fakeSystem) Stat(path string) (os.FileInfo, error) { return os.Stat(path) }
func (f fakeSystem) LookupEnv(key string) (string, bool) {
	value, ok := f.env[key]
	return value, ok
}
func (f fakeSystem) LookPath(name string) (string, error) {
	if path, ok := f.tools[name]; ok {
		return path, nil
	}
	return "", os.ErrNotExist
}

func TestInspectWindowsEnvironment(t *testing.T) {
	home := t.TempDir()
	sdk := filepath.Join(home, "Documents", "PlaydateSDK")
	makeSDK(t, sdk, "windows", "3.1.1", "pdc", "pdutil", "PlaydateSimulator")

	report, err := inspect(context.Background(), fakeSystem{
		goos:   "windows",
		goarch: "amd64",
		home:   home,
		env:    map[string]string{},
		tools: map[string]string{
			"go":  `C:\Go\bin\go.exe`,
			"gcc": `C:\mingw\bin\gcc.exe`,
		},
	}, Config{})
	if err != nil {
		t.Fatal(err)
	}

	if report.Host != "windows/amd64" {
		t.Fatalf("Host = %q, want windows/amd64", report.Host)
	}
	if report.SDKPath != sdk {
		t.Fatalf("SDKPath = %q, want %q", report.SDKPath, sdk)
	}
	if report.SDKVersion != "3.1.1" {
		t.Fatalf("SDKVersion = %q, want 3.1.1", report.SDKVersion)
	}
	assertCapability(t, report, "sdk", StatusReady)
	assertCapability(t, report, "develop", StatusReady)
	assertCapability(t, report, "simulator", StatusUnverified)
	assertCapability(t, report, "device-build", StatusMissing)
	assertCapability(t, report, "device-deploy", StatusUnverified)
}

func TestExplicitSDKPathDoesNotSilentlyFallBack(t *testing.T) {
	home := t.TempDir()
	defaultSDK := filepath.Join(home, "Documents", "PlaydateSDK")
	makeSDK(t, defaultSDK, "windows", "3.1.1", "pdc")

	report, err := inspect(context.Background(), fakeSystem{
		goos:   "windows",
		goarch: "amd64",
		home:   home,
		env:    map[string]string{},
		tools:  map[string]string{"go": `C:\Go\bin\go.exe`},
	}, Config{SDKPath: filepath.Join(home, "does-not-exist")})
	if err != nil {
		t.Fatal(err)
	}

	if report.SDKPath != "" {
		t.Fatalf("SDKPath = %q, want empty for invalid explicit path", report.SDKPath)
	}
	assertCapability(t, report, "sdk", StatusMissing)
}

func TestEnvironmentSDKPathPrecedesConventionalPath(t *testing.T) {
	home := t.TempDir()
	envSDK := filepath.Join(home, "custom-sdk")
	defaultSDK := filepath.Join(home, "Developer", "PlaydateSDK")
	makeSDK(t, envSDK, "linux", "3.2.0", "pdc")
	makeSDK(t, defaultSDK, "linux", "3.1.1", "pdc")

	report, err := inspect(context.Background(), fakeSystem{
		goos:   "linux",
		goarch: "amd64",
		home:   home,
		env:    map[string]string{"PLAYDATE_SDK_PATH": envSDK},
		tools:  map[string]string{"go": "/usr/bin/go"},
	}, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if report.SDKPath != envSDK || report.SDKVersion != "3.2.0" {
		t.Fatalf("selected SDK = %q %q, want environment SDK", report.SDKPath, report.SDKVersion)
	}
	assertCapability(t, report, "sdk", StatusUnverified)
}

func makeSDK(t *testing.T, root, goos, version string, tools ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION.txt"), []byte(version+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools {
		path := filepath.Join(root, "bin", executableName(goos, tool))
		if err := os.WriteFile(path, []byte("test executable"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func assertCapability(t *testing.T, report Report, name string, want Status) {
	t.Helper()
	for _, capability := range report.Capabilities {
		if capability.Name == name {
			if capability.Status != want {
				t.Fatalf("capability %q = %q (%s), want %q", name, capability.Status, capability.Summary, want)
			}
			return
		}
	}
	t.Fatalf("capability %q not found", name)
}
