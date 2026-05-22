package scan

import (
	"testing"
	"time"
)

func TestComputeDelta(t *testing.T) {
	now := time.Now()

	prev := &Snapshot{
		Timestamp: now.Add(-24 * time.Hour),
		Entries: []snapshotEntry{
			{Path: "/Users/test/.npm", Label: "npm cache", SizeMB: 2000},
			{Path: "/Users/test/.Trash", Label: "Trash", SizeMB: 500},
			{Path: "/Users/test/Downloads", Label: "Downloads", SizeMB: 3000},
		},
	}

	curr := []Entry{
		{Path: "/Users/test/.npm", Label: "npm cache", SizeMB: 4100},
		{Path: "/Users/test/.Trash", Label: "Trash", SizeMB: 100},
		{Path: "/Users/test/.cache", Label: "System caches", SizeMB: 1200},
	}

	delta := ComputeDelta(prev, curr)

	if delta.ChangeMB != -100 {
		t.Errorf("expected ChangeMB=-100, got %d", delta.ChangeMB)
	}

	wantItems := 4
	if len(delta.Items) != wantItems {
		t.Errorf("expected %d items, got %d", wantItems, len(delta.Items))
	}

	hasNew := false
	hasGone := false
	hasGrew := false
	hasShrunk := false
	for _, item := range delta.Items {
		switch {
		case item.IsNew:
			hasNew = true
		case item.IsGone:
			hasGone = true
		case item.ChangeMB > 0:
			hasGrew = true
		case item.ChangeMB < 0 && !item.IsGone:
			hasShrunk = true
		}
	}
	if !hasNew {
		t.Error("expected a 'new' item")
	}
	if !hasGone {
		t.Error("expected a 'gone' item")
	}
	if !hasGrew {
		t.Error("expected a 'grew' item")
	}
	if !hasShrunk {
		t.Error("expected a 'shrunk' item")
	}
}

func TestFormatSizeNegative(t *testing.T) {
	tests := []struct {
		mb   int64
		want string
	}{
		{500, "500 MB"},
		{-500, "-500 MB"},
		{2048, "2.0 GB"},
		{-2048, "-2.0 GB"},
		{0, "0 MB"},
	}
	for _, tt := range tests {
		got := FormatSize(tt.mb)
		if got != tt.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.mb, got, tt.want)
		}
	}
}

func TestComputeDeltaEmpty(t *testing.T) {
	now := time.Now()

	prev := &Snapshot{
		Timestamp: now.Add(-1 * time.Hour),
		Entries:   []snapshotEntry{},
	}

	curr := []Entry{}

	delta := ComputeDelta(prev, curr)

	if delta.ChangeMB != 0 {
		t.Errorf("expected ChangeMB=0, got %d", delta.ChangeMB)
	}
	if len(delta.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(delta.Items))
	}
}
