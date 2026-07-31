package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type system interface {
	GOOS() string
	GOARCH() string
	HomeDir() (string, error)
	LookupEnv(string) (string, bool)
	LookPath(string) (string, error)
	ReadFile(string) ([]byte, error)
	Stat(string) (os.FileInfo, error)
}

type hostSystem struct{}

func (hostSystem) GOOS() string                          { return runtime.GOOS }
func (hostSystem) GOARCH() string                        { return runtime.GOARCH }
func (hostSystem) HomeDir() (string, error)              { return os.UserHomeDir() }
func (hostSystem) LookupEnv(key string) (string, bool)   { return os.LookupEnv(key) }
func (hostSystem) LookPath(name string) (string, error)  { return exec.LookPath(name) }
func (hostSystem) ReadFile(path string) ([]byte, error)  { return os.ReadFile(path) }
func (hostSystem) Stat(path string) (os.FileInfo, error) { return os.Stat(path) }

func executableName(goos, name string) string {
	if goos == "windows" {
		return name + ".exe"
	}
	return name
}

func sdkToolPath(goos, root, name string) string {
	return filepath.Join(root, "bin", executableName(goos, name))
}

func simulatorPath(goos, root string) string {
	switch goos {
	case "darwin":
		return filepath.Join(root, "bin", "Playdate Simulator.app", "Contents", "MacOS", "Playdate Simulator")
	default:
		return sdkToolPath(goos, root, "PlaydateSimulator")
	}
}
