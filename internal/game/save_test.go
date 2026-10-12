package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveScoreUsesAppPrivateDirectory(t *testing.T) {
	previous := scoreDirectory
	defer SetScoreDirectory(previous)
	dir := t.TempDir()
	SetScoreDirectory(dir)
	wantPath := filepath.Join(dir, "fruit-fury", "score.json")
	if got := scoreFile(); got != wantPath {
		t.Fatalf("scoreFile() = %q, want %q", got, wantPath)
	}
	saveBestScore(456)
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("score not persisted: %v", err)
	}
	if got := loadBestScore(); got != 456 {
		t.Fatalf("loadBestScore() = %d, want 456", got)
	}
}
