//go:build minimal

package operations

import (
	"context"
	"errors"

	"github.com/rclone/rclone/fs"
)

type listJSONCipher interface {
	EncryptDirName(string) string
	EncryptFileName(string) string
}

func (lj *listJSON) initEncrypted(context.Context, fs.Fs) error {
	if lj.opt.ShowEncrypted {
		return errors.New("lsjson --encrypted is not available in minimal builds")
	}
	return nil
}
