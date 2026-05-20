//go:build !minimal

package operations

import (
	"context"
	"errors"
	"fmt"

	"github.com/rclone/rclone/backend/crypt"
	"github.com/rclone/rclone/fs"
)

type listJSONCipher interface {
	EncryptDirName(string) string
	EncryptFileName(string) string
}

func (lj *listJSON) initEncrypted(ctx context.Context, fsrc fs.Fs) error {
	if !lj.opt.ShowEncrypted {
		return nil
	}
	fsInfo, _, _, config, err := fs.ConfigFs(fs.ConfigStringFull(fsrc))
	if err != nil {
		return fmt.Errorf("ListJSON failed to load config for crypt remote: %w", err)
	}
	if fsInfo.Name != "crypt" {
		return errors.New("the remote needs to be of type \"crypt\"")
	}
	lj.cipher, err = crypt.NewCipher(config)
	if err != nil {
		return fmt.Errorf("ListJSON failed to make new crypt remote: %w", err)
	}
	return nil
}
