package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDownloadActivity(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"Downloading depot 123", true},
		{"Download started for app 456", true},
		{"Download complete", true},
		{"Download finished", true},
		{"Update required for app 456", true},
		{"Update queued", true},
		{"Update started", true},
		{"Installing game files", true},
		{"Patching content", true},
		{"AppUpdate for 123", true},
		{"Depot download in progress", true},
		{"Staging files", true},
		{"Validating installation", true},
		{"Preallocating disk space", true},
		{"Steam client started", false},
		{"User logged in", false},
		{"Cloud sync complete", false},
		{"Game: The Last Update", false}, // "update" in game name shouldn't match
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isDownloadActivity(tt.line)
			if got != tt.want {
				t.Errorf("isDownloadActivity(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"123", true},
		{"0", true},
		{"999999", true},
		{"", false},
		{"12a3", false},
		{"abc", false},
		{"-1", false},
		{"1.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := isNumeric(tt.s)
			if got != tt.want {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestDirHasEntries(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		got, err := dirHasEntries(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != false {
			t.Errorf("dirHasEntries(%q) = %v, want false", dir, got)
		}
	})

	t.Run("directory with file", func(t *testing.T) {
		dir := t.TempDir()
		f, err := os.Create(filepath.Join(dir, "test.txt"))
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		f.Close()

		got, err := dirHasEntries(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != true {
			t.Errorf("dirHasEntries(%q) = %v, want true", dir, got)
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := dirHasEntries("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})
}

func TestCollectSubdirs(t *testing.T) {
	steamapps := []string{"/path/to/steamapps", "/other/steamapps"}
	got := collectSubdirs(steamapps, "downloading")
	want := []string{"/path/to/steamapps/downloading", "/other/steamapps/downloading"}

	if len(got) != len(want) {
		t.Fatalf("collectSubdirs() returned %d items, want %d", len(got), len(want))
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("collectSubdirs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
