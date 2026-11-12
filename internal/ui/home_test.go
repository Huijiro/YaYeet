package ui

import (
	"testing"

	"github.com/Huijiro/YaYeet/internal/game"
)

func TestPreferredVersionLabelPrefersInstalled(t *testing.T) {
	byLabel := map[string]game.VersionOption{
		"0.9.0 (latest)": {Name: "0.9.0", Label: "0.9.0 (latest)"},
		"0.7.5":          {Name: "0.7.5", Label: "0.7.5"},
	}

	got := preferredVersionLabel(byLabel, "0.7.5", "0.9.0 (latest)")
	if got != "0.7.5" {
		t.Fatalf("preferredVersionLabel() = %q, want installed 0.7.5", got)
	}
}

func TestPreferredVersionLabelFallsBackToLatest(t *testing.T) {
	byLabel := map[string]game.VersionOption{
		"0.9.0 (latest)": {Name: "0.9.0", Label: "0.9.0 (latest)"},
		"0.7.5":          {Name: "0.7.5", Label: "0.7.5"},
	}

	got := preferredVersionLabel(byLabel, "", "0.9.0 (latest)")
	if got != "0.9.0 (latest)" {
		t.Fatalf("preferredVersionLabel() = %q, want latest", got)
	}

	got = preferredVersionLabel(byLabel, "missing", "0.9.0 (latest)")
	if got != "0.9.0 (latest)" {
		t.Fatalf("preferredVersionLabel(missing) = %q, want latest", got)
	}
}
