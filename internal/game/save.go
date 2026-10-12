package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type scoreData struct {
	Best int `json:"best"`
}

// Android supplies its private files directory from the host Activity.
// No storage permissions are needed to save scores in this directory.
var scoreDirectory string

func SetScoreDirectory(directory string) {
	scoreDirectory = directory
}

func scoreFile() string {
	if scoreDirectory != "" {
		return filepath.Join(scoreDirectory, "fruit-fury", "score.json")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "fruit-fury", "score.json")
}

func loadBestScore() int {
	p := scoreFile()
	if p == "" {
		return 0
	}
	bytes, err := os.ReadFile(p)
	if err != nil {
		return 0
	}
	var data scoreData
	if json.Unmarshal(bytes, &data) != nil || data.Best < 0 {
		return 0
	}
	return data.Best
}

func saveBestScore(best int) {
	p := scoreFile()
	if p == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return
	}
	bytes, err := json.MarshalIndent(scoreData{Best: best}, "", "  ")
	if err != nil {
		return
	}
	// A temporary file prevents partially written high scores after a crash.
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, bytes, 0644); err != nil {
		return
	}
	if err := os.Rename(tmp, p); err != nil {
		// Some Windows filesystems cannot rename over an existing file.
		_ = os.Remove(p)
		if err := os.Rename(tmp, p); err != nil {
			_ = os.Remove(tmp)
		}
	}
}
