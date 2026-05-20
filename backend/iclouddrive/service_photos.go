//go:build !minimal

package iclouddrive

import (
	"context"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
)

func serviceDescription() string {
	return "iCloud Drive and Photos"
}

func serviceMetadataInfo() *fs.MetadataInfo {
	return &fs.MetadataInfo{
		System: map[string]fs.MetadataHelp{
			"width": {
				Help:     "Image width in pixels",
				Type:     "int",
				ReadOnly: true,
			},
			"height": {
				Help:     "Image height in pixels",
				Type:     "int",
				ReadOnly: true,
			},
			"added-time": {
				Help:     "Time the item was added to the iCloud library",
				Type:     "RFC 3339",
				Example:  "2006-01-02T15:04:05Z",
				ReadOnly: true,
			},
			"favorite": {
				Help:     "Whether the item is marked as favorite",
				Type:     "bool",
				ReadOnly: true,
			},
			"hidden": {
				Help:     "Whether the item is hidden",
				Type:     "bool",
				ReadOnly: true,
			},
		},
		Help: "Metadata is read-only and available for the Photos service only.",
	}
}

func serviceOptionExamples() []fs.OptionExample {
	return []fs.OptionExample{{
		Value: serviceDrive,
		Help:  "iCloud Drive",
	}, {
		Value: servicePhotos,
		Help:  "iCloud Photos",
	}}
}

func serviceSelectionHelp() string {
	return "must be 'drive' or 'photos'"
}

func newServiceFsPhotos(ctx context.Context, name, root string, m configmap.Mapper) (fs.Fs, error) {
	return NewFsPhotos(ctx, name, root, m)
}
