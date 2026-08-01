package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIExternalConsumerWorkflow(t *testing.T) {
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	binaryName := "gopdsdk"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binary := filepath.Join(t.TempDir(), binaryName)
	runTestCommand(t, repository, "go", "build", "-buildvcs=false", "-o", binary, "./cmd/gopdsdk")

	project := filepath.Join(t.TempDir(), "external game")
	runTestCommand(t, repository, binary, "init", "--module", "example.com/acceptance", "--author", "CI", "--bundle-id", "com.example.acceptance", project)
	goMod, err := os.ReadFile(filepath.Join(project, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "replace github.com/Djunichi/gopdsdk =>") {
		t.Fatalf("checkout acceptance go.mod does not contain a local replace:\n%s", goMod)
	}
	runTestCommand(t, project, "go", "mod", "tidy")
	runTestCommand(t, project, "go", "test", "./...")

	for _, test := range []struct {
		arguments []string
		target    string
	}{
		{arguments: []string{"build", "--dry-run", "--sdk", filepath.Join(project, "fake sdk"), "."}, target: "simulator"},
		{arguments: []string{"build", "device", "--dry-run", "--sdk", filepath.Join(project, "fake sdk"), "."}, target: "device"},
	} {
		output := runTestCommand(t, project, binary, test.arguments...)
		for _, required := range []string{"Build plan", "Target:      " + test.target, "Application: ."} {
			if !strings.Contains(output, required) {
				t.Fatalf("%v output does not contain %q:\n%s", test.arguments, required, output)
			}
		}
	}
}

func TestCLIPlayableExternalConsumerWorkflow(t *testing.T) {
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	binaryName := "gopdsdk"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binary := filepath.Join(t.TempDir(), binaryName)
	runTestCommand(t, repository, "go", "build", "-buildvcs=false", "-o", binary, "./cmd/gopdsdk")

	project := filepath.Join(t.TempDir(), "p1.3 playable consumer")
	copyTestTree(t, filepath.Join(repository, "cmd", "gopdsdk", "testdata", "playable"), project)
	module := fmt.Sprintf("module example.com/p13-playable\n\ngo 1.26.5\n\nrequire github.com/Djunichi/gopdsdk v0.1.0\n\nreplace github.com/Djunichi/gopdsdk => %q\n", filepath.ToSlash(repository))
	if err := os.WriteFile(filepath.Join(project, "go.mod"), []byte(module), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, project, "go", "test", "./...")

	for _, test := range []struct {
		arguments []string
		target    string
	}{
		{arguments: []string{"build", "--dry-run", "--sdk", filepath.Join(project, "fake sdk"), "."}, target: "simulator"},
		{arguments: []string{"build", "device", "--dry-run", "--sdk", filepath.Join(project, "fake sdk"), "."}, target: "device"},
	} {
		output := runTestCommand(t, project, binary, test.arguments...)
		for _, required := range []string{"Build plan", "Target:      " + test.target, "Application: ."} {
			if !strings.Contains(output, required) {
				t.Fatalf("%v output does not contain %q:\n%s", test.arguments, required, output)
			}
		}
	}
}

func copyTestTree(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func copyFile(source, destination string) error {
	contents, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, contents, 0o644)
}
func runTestCommand(t *testing.T, directory, executable string, arguments ...string) string {
	t.Helper()
	command := exec.Command(executable, arguments...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", executable, arguments, err, output)
	}
	return string(output)
}
