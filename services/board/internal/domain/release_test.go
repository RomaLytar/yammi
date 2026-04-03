package domain

import (
	"errors"
	"testing"
)

func TestNewRelease(t *testing.T) {
	t.Parallel()

	t.Run("valid release", func(t *testing.T) {
		r, err := NewRelease("board-1", "Sprint 1", "First sprint", "user-1", nil, nil)
		if err != nil {
			t.Fatalf("NewRelease() error = %v", err)
		}
		if r.ID == "" {
			t.Error("ID should be generated")
		}
		if r.BoardID != "board-1" {
			t.Errorf("BoardID = %q, want %q", r.BoardID, "board-1")
		}
		if r.Name != "Sprint 1" {
			t.Errorf("Name = %q, want %q", r.Name, "Sprint 1")
		}
		if r.Status != ReleaseStatusDraft {
			t.Errorf("Status = %q, want %q", r.Status, ReleaseStatusDraft)
		}
		if r.Version != 1 {
			t.Errorf("Version = %d, want 1", r.Version)
		}
		if r.StartedAt != nil {
			t.Error("StartedAt should be nil for draft")
		}
		if r.CompletedAt != nil {
			t.Error("CompletedAt should be nil for draft")
		}
	})

	t.Run("empty board ID", func(t *testing.T) {
		_, err := NewRelease("", "Sprint 1", "", "user-1", nil, nil)
		if !errors.Is(err, ErrBoardNotFound) {
			t.Errorf("error = %v, want ErrBoardNotFound", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewRelease("board-1", "", "", "user-1", nil, nil)
		if !errors.Is(err, ErrEmptyReleaseName) {
			t.Errorf("error = %v, want ErrEmptyReleaseName", err)
		}
	})

	t.Run("empty creator", func(t *testing.T) {
		_, err := NewRelease("board-1", "Sprint 1", "", "", nil, nil)
		if !errors.Is(err, ErrEmptyActorID) {
			t.Errorf("error = %v, want ErrEmptyActorID", err)
		}
	})
}

func TestRelease_Update(t *testing.T) {
	t.Parallel()

	t.Run("valid update", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "desc", "user-1", nil, nil)
		err := r.Update("Sprint 1 updated", "new desc", nil, nil)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if r.Name != "Sprint 1 updated" {
			t.Errorf("Name = %q, want %q", r.Name, "Sprint 1 updated")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		err := r.Update("", "desc", nil, nil)
		if !errors.Is(err, ErrEmptyReleaseName) {
			t.Errorf("error = %v, want ErrEmptyReleaseName", err)
		}
	})

	t.Run("completed release immutable", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		_ = r.Start(14)
		_ = r.Complete()
		err := r.Update("new name", "new desc", nil, nil)
		if !errors.Is(err, ErrReleaseCompleted) {
			t.Errorf("error = %v, want ErrReleaseCompleted", err)
		}
	})
}

func TestRelease_Start(t *testing.T) {
	t.Parallel()

	t.Run("draft to active", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		err := r.Start(14)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if r.Status != ReleaseStatusActive {
			t.Errorf("Status = %q, want %q", r.Status, ReleaseStatusActive)
		}
		if r.StartedAt == nil {
			t.Error("StartedAt should be set")
		}
	})

	t.Run("active cannot start again", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		_ = r.Start(14)
		err := r.Start(14)
		if !errors.Is(err, ErrReleaseNotDraft) {
			t.Errorf("error = %v, want ErrReleaseNotDraft", err)
		}
	})

	t.Run("completed cannot start", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		_ = r.Start(14)
		_ = r.Complete()
		err := r.Start(14)
		if !errors.Is(err, ErrReleaseNotDraft) {
			t.Errorf("error = %v, want ErrReleaseNotDraft", err)
		}
	})
}

func TestRelease_Complete(t *testing.T) {
	t.Parallel()

	t.Run("active to completed", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		_ = r.Start(14)
		err := r.Complete()
		if err != nil {
			t.Fatalf("Complete() error = %v", err)
		}
		if r.Status != ReleaseStatusCompleted {
			t.Errorf("Status = %q, want %q", r.Status, ReleaseStatusCompleted)
		}
		if r.CompletedAt == nil {
			t.Error("CompletedAt should be set")
		}
	})

	t.Run("draft cannot complete", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		err := r.Complete()
		if !errors.Is(err, ErrReleaseNotActive) {
			t.Errorf("error = %v, want ErrReleaseNotActive", err)
		}
	})

	t.Run("completed cannot complete again", func(t *testing.T) {
		r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
		_ = r.Start(14)
		_ = r.Complete()
		err := r.Complete()
		if !errors.Is(err, ErrReleaseNotActive) {
			t.Errorf("error = %v, want ErrReleaseNotActive", err)
		}
	})
}

func TestRelease_StatusHelpers(t *testing.T) {
	t.Parallel()

	r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)

	if !r.IsDraft() {
		t.Error("new release should be draft")
	}
	if r.IsActive() || r.IsCompleted() {
		t.Error("new release should not be active or completed")
	}

	_ = r.Start(14)
	if !r.IsActive() {
		t.Error("started release should be active")
	}
	if r.IsDraft() || r.IsCompleted() {
		t.Error("started release should not be draft or completed")
	}

	_ = r.Complete()
	if !r.IsCompleted() {
		t.Error("completed release should be completed")
	}
	if r.IsDraft() || r.IsActive() {
		t.Error("completed release should not be draft or active")
	}
}

func TestRelease_IncrementVersion(t *testing.T) {
	t.Parallel()

	r, _ := NewRelease("board-1", "Sprint 1", "", "user-1", nil, nil)
	if r.Version != 1 {
		t.Fatalf("initial version = %d, want 1", r.Version)
	}
	r.IncrementVersion()
	if r.Version != 2 {
		t.Errorf("version after increment = %d, want 2", r.Version)
	}
}

func TestReleaseStatus_IsValid(t *testing.T) {
	t.Parallel()

	valid := []ReleaseStatus{ReleaseStatusDraft, ReleaseStatusActive, ReleaseStatusCompleted}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("%q should be valid", s)
		}
	}

	if ReleaseStatus("unknown").IsValid() {
		t.Error("unknown should be invalid")
	}
}
