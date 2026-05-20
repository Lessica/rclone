//go:build minimal

package bisync

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/rclone/rclone/cmd/bisync/bilib"
	"github.com/rclone/rclone/cmd/check"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/accounting"
	"github.com/rclone/rclone/fs/filter"
	"github.com/rclone/rclone/fs/hash"
	"github.com/rclone/rclone/fs/operations"
)

type bisyncCheck = struct {
	hashType   hash.Type
	fsrc, fdst fs.Fs
}

// WhichCheck determines which CheckFn we should use based on the Fs types.
func (b *bisyncRun) WhichCheck(ctx context.Context, opt *operations.CheckOpt) *operations.CheckOpt {
	ci := fs.GetConfig(ctx)
	common := opt.Fsrc.Hashes().Overlap(opt.Fdst.Hashes())

	if common.Count() > 0 || ci.SizeOnly || ci.IgnoreChecksum {
		opt.Check = CheckFn
		return opt
	}

	fs.Infof(opt.Fdst, "Can't compare hashes, so using check --download for safety. (Use --size-only or --ignore-checksum to disable)")
	opt.Check = DownloadCheckFn
	return opt
}

// CheckFn is a slightly modified version of Check.
func CheckFn(ctx context.Context, dst, src fs.Object) (differ bool, noHash bool, err error) {
	same, ht, err := operations.CheckHashes(ctx, src, dst)
	if err != nil {
		return true, false, err
	}
	if ht == hash.None {
		return false, true, nil
	}
	if !same {
		err = fmt.Errorf("%v differ", ht)
		fs.Errorf(src, "%v", err)
		return true, false, nil
	}
	return false, false, nil
}

// DownloadCheckFn is a slightly modified version of Check with --download.
func DownloadCheckFn(ctx context.Context, dst, src fs.Object) (equal bool, noHash bool, err error) {
	equal, err = operations.CheckIdenticalDownload(ctx, src, dst)
	if err != nil {
		return true, true, fmt.Errorf("failed to download: %w", err)
	}
	return !equal, false, nil
}

// check potential conflicts (to avoid renaming if already identical)
func (b *bisyncRun) checkconflicts(ctxCheck context.Context, filterCheck *filter.Filter, fs1, fs2 fs.Fs) (bilib.Names, error) {
	matches := bilib.Names{}
	if filterCheck.HaveFilesFrom() {
		fs.Debugf(nil, "There are potential conflicts to check.")

		opt, close, checkopterr := check.GetCheckOpt(fs1, fs2)
		if checkopterr != nil {
			b.critical = true
			b.retryable = true
			fs.Debugf(nil, "GetCheckOpt error: %v", checkopterr)
			return matches, checkopterr
		}
		defer close()

		opt.Match = new(bytes.Buffer)
		opt = b.WhichCheck(ctxCheck, opt)

		fs.Infof(nil, "Checking potential conflicts...")
		check := operations.CheckFn(ctxCheck, opt)
		fs.Infof(nil, "Finished checking the potential conflicts. %s", check)

		accounting.Stats(ctxCheck).ResetErrors()

		if len(fmt.Sprint(opt.Match)) > 0 {
			matches = bilib.ToNames(strings.Split(fmt.Sprint(opt.Match), "\n"))
		}
		if matches.NotEmpty() {
			fs.Debugf(nil, "The following potential conflicts were determined to be identical. %v", matches)
		} else {
			fs.Debugf(nil, "None of the conflicts were determined to be identical.")
		}
	}
	return matches, nil
}

// WhichEqual is similar to WhichCheck, but checks a single object.
func (b *bisyncRun) WhichEqual(ctx context.Context, src, dst fs.Object, Fsrc, Fdst fs.Fs) bool {
	opt, close, checkopterr := check.GetCheckOpt(Fsrc, Fdst)
	if checkopterr != nil {
		fs.Debugf(nil, "GetCheckOpt error: %v", checkopterr)
	}
	defer close()

	opt = b.WhichCheck(ctx, opt)
	differ, noHash, err := opt.Check(ctx, dst, src)
	if err != nil {
		fs.Errorf(src, "failed to check: %v", err)
		return false
	}
	if noHash {
		fs.Errorf(src, "failed to check as hash is missing")
		return false
	}
	return !differ
}

// EqualFn replaces the standard Equal func with one that also considers checksum.
func (b *bisyncRun) EqualFn(ctx context.Context) context.Context {
	ci := fs.GetConfig(ctx)
	ci.CheckSum = false
	var equalFn operations.EqualFn = func(ctx context.Context, src fs.ObjectInfo, dst fs.Object) bool {
		fs.Debugf(src, "evaluating...")
		logger, _ := operations.GetLogger(ctx)
		noop := func(ctx context.Context, sigil operations.Sigil, src, dst fs.DirEntry, err error) {
			fs.Debugf(src, "equal skipped")
		}
		ctxNoLogger := operations.WithLogger(ctx, noop)

		timeSizeEqualFn := func() (equal bool, skipHash bool) { return operations.Equal(ctxNoLogger, src, dst), false }
		if b.opt.ResyncMode == PreferOlder || b.opt.ResyncMode == PreferLarger || b.opt.ResyncMode == PreferSmaller {
			timeSizeEqualFn = func() (equal bool, skipHash bool) { return b.resyncTimeSizeEqual(ctxNoLogger, src, dst) }
		}
		skipHash := false
		equal, skipHash := timeSizeEqualFn()
		if equal && !skipHash {
			whichHashType := func(f fs.Info) hash.Type {
				ht := b.getHashType(f.Name())
				if ht == hash.None && b.opt.Compare.SlowHashSyncOnly && !b.opt.Resync {
					ht = f.Hashes().GetOne()
				}
				return ht
			}
			srcHash, _ := src.Hash(ctx, whichHashType(src.Fs()))
			dstHash, _ := dst.Hash(ctx, whichHashType(dst.Fs()))
			srcHash, _ = b.tryDownloadHash(ctx, src, srcHash)
			dstHash, _ = b.tryDownloadHash(ctx, dst, dstHash)
			equal = !b.hashDiffers(srcHash, dstHash, whichHashType(src.Fs()), whichHashType(dst.Fs()), src.Size(), dst.Size())
		}
		if equal {
			logger(ctx, operations.Match, src, dst, nil)
			fs.Debugf(src, "EqualFn: files are equal")
			return true
		}
		logger(ctx, operations.Differ, src, dst, nil)
		fs.Debugf(src, "EqualFn: files are NOT equal")
		return false
	}
	return operations.WithEqualFn(ctx, equalFn)
}

func (b *bisyncRun) resyncTimeSizeEqual(ctxNoLogger context.Context, src fs.ObjectInfo, dst fs.Object) (equal bool, skipHash bool) {
	switch b.opt.ResyncMode {
	case PreferLarger, PreferSmaller:
		path1, path2 := b.resyncWhichIsWhich(src, dst)
		if sizeDiffers(path1.Size(), path2.Size()) {
			winningPath := b.resolveLargerSmaller(path1.Size(), path2.Size(), path1.Remote(), b.opt.ResyncMode)
			return b.resyncWinningPathToEqual(winningPath), b.resyncWinningPathToEqual(winningPath)
		}
		return operations.Equal(ctxNoLogger, src, dst), false
	case PreferOlder:
		path1, path2 := b.resyncWhichIsWhich(src, dst)
		if timeDiffers(ctxNoLogger, path1.ModTime(ctxNoLogger), path2.ModTime(ctxNoLogger), path1.Fs(), path2.Fs()) {
			winningPath := b.resolveNewerOlder(path1.ModTime(ctxNoLogger), path2.ModTime(ctxNoLogger), path1.Remote(), b.opt.ResyncMode)
			if !b.resyncWinningPathToEqual(winningPath) {
				return operations.Equal(ctxNoLogger, src, dst), false
			}
			return true, true
		}
	}
	return operations.Equal(ctxNoLogger, src, dst), false
}
