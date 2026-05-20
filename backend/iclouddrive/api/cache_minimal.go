//go:build minimal

package api

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/rclone/rclone/fs"
)

const cacheSubdir = "iclouddrive-photos"

func saveJSONCache(dir, filename string, v any) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		fs.Debugf(nil, "iclouddrive: failed to create cache dir: %v", err)
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		fs.Debugf(nil, "iclouddrive: failed to marshal cache %s: %v", filename, err)
		return
	}
	if err := atomicWriteFile(filepath.Join(dir, filename), data); err != nil {
		fs.Debugf(nil, "iclouddrive: failed to write cache %s: %v", filename, err)
	}
}

func atomicWriteFile(target string, data []byte) error {
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
