package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsWatcher_initWatcher(t *testing.T) {
	fsWatcher := FsWatcher{}

	// initialize fsWatcher when the fsnotify watcher is nil
	err := fsWatcher.initWatcher()
	assert.NoError(t, err)

	// initialize fsWatcher when the fsnotify watcher is not nil nil
	err = fsWatcher.initWatcher()
	assert.NoError(t, err)
}

func TestFsWatcher_add(t *testing.T) {
	f, err := os.Create("NannyConfiguration")
	defer os.Remove(f.Name())
	assert.NoError(t, err)

	configDir, err := filepath.Abs(filepath.Dir(f.Name()))
	assert.NoError(t, err)

	configFile := filepath.Join(configDir, "NannyConfiguration")
	fsWatcher := FsWatcher{}

	err = fsWatcher.initWatcher()
	assert.NoError(t, err)

	err = fsWatcher.add(configFile)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(fsWatcher.paths), 1)
}
