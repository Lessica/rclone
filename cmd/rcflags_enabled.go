//go:build !minimal

package cmd

import (
	"github.com/rclone/rclone/fs/rc/rcflags"
	"github.com/spf13/pflag"
)

func addRCFlags(flagSet *pflag.FlagSet) {
	rcflags.AddFlags(flagSet)
}
