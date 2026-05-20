//go:build minimal

package iclouddrive

import (
	"context"
	"errors"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
)

func serviceDescription() string {
	return "iCloud Drive"
}

func serviceMetadataInfo() *fs.MetadataInfo {
	return nil
}

func serviceOptionExamples() []fs.OptionExample {
	return []fs.OptionExample{{
		Value: serviceDrive,
		Help:  "iCloud Drive",
	}}
}

func serviceSelectionHelp() string {
	return "must be 'drive'"
}

func newServiceFsPhotos(context.Context, string, string, configmap.Mapper) (fs.Fs, error) {
	return nil, errors.New("iCloud Photos is not available in minimal builds")
}
