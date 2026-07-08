package scan

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverHomeUsesManualTier(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, "BigFolder")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFileOfSize(t, filepath.Join(target, "blob.bin"), 120)

	entries := discoverHome(home, map[string]bool{})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Tier != TierManual {
		t.Fatalf("expected manual tier, got %d", entries[0].Tier)
	}
	if entries[0].Cleanable {
		t.Fatal("expected manual entry to be non-cleanable")
	}
}

func TestDiscoverAppSupportSkipsAncestorOfKnownTarget(t *testing.T) {
	home := t.TempDir()
	parent := filepath.Join(home, "Library", "Application Support", "Google")
	child := filepath.Join(parent, "Chrome")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFileOfSize(t, filepath.Join(child, "profile.dat"), 600)

	known := map[string]bool{
		"Library/Application Support/Google/Chrome/": true,
	}
	entries := discoverAppSupport(home, known)
	if len(entries) != 0 {
		t.Fatalf("expected no ancestor discovery, got %d entries", len(entries))
	}
}

func TestDiscoverAppSupportKeepsParentWhenKnownChildMissing(t *testing.T) {
	home := t.TempDir()
	parent := filepath.Join(home, "Library", "Application Support", "Google")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFileOfSize(t, filepath.Join(parent, "cache.bin"), 600)

	known := map[string]bool{
		"Library/Application Support/Google/Chrome/": true,
	}
	entries := discoverAppSupport(home, known)
	if len(entries) != 1 {
		t.Fatalf("expected parent to be discovered, got %d entries", len(entries))
	}
}

func TestDiscoverTelegramFindsMatchingGroupContainer(t *testing.T) {
	home := t.TempDir()
	parent := filepath.Join(home, "Library", "Group Containers", "6N38VWS5BX.ru.keepcoder.Telegram")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFileOfSize(t, filepath.Join(parent, "media.bin"), 250)

	entries := discoverGroupContainers(home)
	if len(entries) != 1 {
		t.Fatalf("expected 1 telegram entry, got %d", len(entries))
	}
	if entries[0].Label != "Telegram media" {
		t.Fatalf("unexpected label: %s", entries[0].Label)
	}
	if entries[0].Tier != TierApp {
		t.Fatalf("expected tier 3, got %d", entries[0].Tier)
	}
}

func writeFileOfSize(t *testing.T, path string, sizeMB int) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	chunk := bytes.Repeat([]byte("x"), 1024*1024)
	for i := 0; i < sizeMB; i++ {
		if _, err := f.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}
}
