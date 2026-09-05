package buildinfo

import "testing"

func TestUpdateChecksEnabled(t *testing.T) {
	originalVersion := Version
	originalEnabled := UpdatesEnabled
	t.Cleanup(func() {
		Version = originalVersion
		UpdatesEnabled = originalEnabled
	})

	tests := []struct {
		name    string
		version string
		enabled string
		want    bool
	}{
		{name: "release", version: "1.4.0", enabled: "true", want: true},
		{name: "development version", version: "dev", enabled: "true", want: false},
		{name: "disabled release", version: "1.4.0", enabled: "false", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Version = test.version
			UpdatesEnabled = test.enabled
			if got := UpdateChecksEnabled(); got != test.want {
				t.Fatalf("UpdateChecksEnabled() = %v, want %v", got, test.want)
			}
		})
	}
}
