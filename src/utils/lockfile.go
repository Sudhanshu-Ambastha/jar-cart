package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type LockEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type LockFile struct {
	Version      int                  `json:"version"`
	GeneratedAt  string               `json:"generated_at"`
	Dependencies map[string]LockEntry `json:"dependencies"`
}

func LoadLockFile(projectDir string) (*LockFile, error) {
	lockPath := filepath.Join(projectDir, "jar-cart.lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}
	return &lock, nil
}

func WriteLockFile(projectDir string, lock *LockFile) error {
	lockPath := filepath.Join(projectDir, "jar-cart.lock")
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(lockPath, data, 0644)
}

func CalculateSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}