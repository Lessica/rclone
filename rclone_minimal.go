//go:build minimal

// Sync files and directories to and from selected local and remote object stores.
package main

import (
	_ "github.com/rclone/rclone/backend/all"
	"github.com/rclone/rclone/cmd"
	_ "github.com/rclone/rclone/cmd/all"
)

func main() {
	cmd.Main()
}
