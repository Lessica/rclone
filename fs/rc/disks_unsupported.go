//go:build netbsd && 386 && !minimal

package rc

// getMounts returns a slice of disk mount points
func getMounts() (mounts []string) {
	return []string{"/"}
}
