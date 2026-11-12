package ui

import (
	"sync"

	"fyne.io/fyne/v2"
)

// installSession tracks an in-flight version install across home remounts.
// Opening Configuration destroys the home page widgets; this keeps progress
// and single-flight state so the install cannot be lost or started twice.
type installSession struct {
	mu sync.Mutex

	running           bool
	version           string
	progress          float64
	completedVersion  string
	updateUI          func(progress float64)
	finishUI          func(version string, err error)
}

type installSnapshot struct {
	Running          bool
	Version          string
	Progress         float64
	CompletedVersion string
}

func (s *installSession) bind(updateUI func(progress float64), finishUI func(version string, err error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateUI = updateUI
	s.finishUI = finishUI
}

func (s *installSession) unbind() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateUI = nil
	s.finishUI = nil
}

func (s *installSession) snapshot() installSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return installSnapshot{
		Running:          s.running,
		Version:          s.version,
		Progress:         s.progress,
		CompletedVersion: s.completedVersion,
	}
}

func (s *installSession) tryStart(version string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	s.version = version
	s.progress = 0
	return true
}

func (s *installSession) setProgress(progress float64) {
	s.mu.Lock()
	s.progress = progress
	update := s.updateUI
	s.mu.Unlock()
	if update != nil {
		fyne.Do(func() {
			update(progress)
		})
	}
}

func (s *installSession) finish(version string, err error) {
	s.mu.Lock()
	s.running = false
	if err == nil {
		s.progress = 1
		s.completedVersion = version
		s.version = version
	}
	finish := s.finishUI
	s.mu.Unlock()
	if finish != nil {
		fyne.Do(func() {
			finish(version, err)
		})
	}
}
