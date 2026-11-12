package ui

import "testing"

func TestInstallSessionTryStartIsSingleFlight(t *testing.T) {
	session := &installSession{}
	if !session.tryStart("0100_0009") {
		t.Fatal("tryStart() first call = false, want true")
	}
	if session.tryStart("0100_0010") {
		t.Fatal("tryStart() while running = true, want false")
	}

	snapshot := session.snapshot()
	if !snapshot.Running {
		t.Fatal("snapshot.Running = false, want true")
	}
	if snapshot.Version != "0100_0009" {
		t.Fatalf("snapshot.Version = %q, want 0100_0009", snapshot.Version)
	}
}

func TestInstallSessionSurvivesUnbind(t *testing.T) {
	session := &installSession{}
	if !session.tryStart("0100_0009") {
		t.Fatal("tryStart() = false, want true")
	}
	session.setProgress(0.4)
	session.unbind()

	snapshot := session.snapshot()
	if !snapshot.Running {
		t.Fatal("snapshot.Running after unbind = false, want true")
	}
	if snapshot.Progress != 0.4 {
		t.Fatalf("snapshot.Progress = %v, want 0.4", snapshot.Progress)
	}

	session.finish("0100_0009", nil)
	snapshot = session.snapshot()
	if snapshot.Running {
		t.Fatal("snapshot.Running after finish = true, want false")
	}
	if snapshot.CompletedVersion != "0100_0009" {
		t.Fatalf("snapshot.CompletedVersion = %q, want 0100_0009", snapshot.CompletedVersion)
	}
	if !session.tryStart("0100_0010") {
		t.Fatal("tryStart() after finish = false, want true")
	}
}
