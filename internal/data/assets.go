package data

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// OpenAssetsFS locates the repo assets directory from the current working directory
// or one of its parents and returns it as an fs.FS plus its absolute path.
func OpenAssetsFS() (fs.FS, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	dir, err := FindAssetsDir(cwd)
	if err != nil {
		return nil, "", err
	}

	return os.DirFS(dir), dir, nil
}

// FindAssetsDir walks upward from startDir until it finds an assets directory that
// contains data/rooms.json.
func FindAssetsDir(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, "assets")
		if fileExists(filepath.Join(candidate, "data", "rooms.json")) {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not locate assets/data/rooms.json from %s", startDir)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
