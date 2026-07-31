package build

import "testing"

func TestValidatePDXInfo(t *testing.T) {
	contents := []byte("name=Game\nauthor=Author\nbundleID=com.example.game\nversion=1.0\nbuildNumber=1\ndescription=Optional\n")
	if err := validatePDXInfo(contents); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePDXInfoRejectsInvalidMetadata(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{"missing field", "name=Game\n", "field author is required"},
		{"invalid bundle", "name=Game\nauthor=A\nbundleID=game\nversion=1\nbuildNumber=1\n", "field bundleID must use reverse DNS notation"},
		{"invalid build", "name=Game\nauthor=A\nbundleID=com.example.game\nversion=1\nbuildNumber=0\n", "field buildNumber must be a positive integer"},
		{"duplicate", "name=Game\nname=Other\n", "field name is duplicated"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePDXInfo([]byte(test.contents))
			if err == nil || err.Error() != test.want {
				t.Fatalf("validatePDXInfo() error = %v, want %q", err, test.want)
			}
		})
	}
}
