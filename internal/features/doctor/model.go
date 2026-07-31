package doctor

import "fmt"

// Status describes whether a development capability can be used.
type Status string

const (
	StatusReady        Status = "ready"
	StatusMissing      Status = "missing"
	StatusIncompatible Status = "incompatible"
	StatusUnverified   Status = "unverified"
)

// Capability is one independently useful part of the development environment.
type Capability struct {
	Name    string
	Status  Status
	Summary string
}

// Tool is an executable discovered on the host.
type Tool struct {
	Name    string
	Path    string
	Version string
}

// Report is the complete, display-ready environment assessment.
type Report struct {
	Host         string
	SDKPath      string
	SDKVersion   string
	Tools        []Tool
	Capabilities []Capability
}

func (r Report) tool(name string) (Tool, bool) {
	for _, tool := range r.Tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return Tool{}, false
}

func (s Status) valid() bool {
	switch s {
	case StatusReady, StatusMissing, StatusIncompatible, StatusUnverified:
		return true
	default:
		return false
	}
}

func (c Capability) validate() error {
	if c.Name == "" {
		return fmt.Errorf("capability name is empty")
	}
	if !c.Status.valid() {
		return fmt.Errorf("capability %q has invalid status %q", c.Name, c.Status)
	}
	return nil
}
