//go:build !minimal

package cmd

import (
	"context"
	"os"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/rc"
	"github.com/rclone/rclone/fs/rc/rcserver"
)

func startBackgroundServices(ctx context.Context) {
	// Start the remote control server if configured
	_, err := rcserver.Start(ctx, &rc.Opt)
	if err != nil {
		fs.Fatalf(nil, "Failed to start remote control: %v", err)
	}

	// Start the metrics server if configured and not running the "rc" command
	if len(os.Args) >= 2 && os.Args[1] != "rc" {
		_, err = rcserver.MetricsStart(ctx, &rc.Opt)
		if err != nil {
			fs.Fatalf(nil, "Failed to start metrics server: %v", err)
		}
	}
}
