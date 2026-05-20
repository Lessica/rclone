//go:build minimal

package bisync

import (
	"strconv"
	"strings"
)

func addRC() {}

var shortHelp = `Perform bidirectional synchronization between two paths.`

var longHelp = shortHelp + MakeHelp(`

[Bisync](https://rclone.org/bisync/) provides a
bidirectional cloud sync solution in rclone.
It retains the Path1 and Path2 filesystem listings from the prior run.
On each successive run it will:

- list files on Path1 and Path2, and check for changes on each side.
  Changes include ||New||, ||Newer||, ||Older||, and ||Deleted|| files.
- Propagate changes on Path1 to Path2, and vice-versa.

Bisync is considered an **advanced command**, so use with care.
Make sure you have read and understood the entire [manual](https://rclone.org/bisync)
(especially the [Limitations](https://rclone.org/bisync/#limitations) section)
before using, or data loss can result. Questions can be asked in the
[Rclone Forum](https://forum.rclone.org/).

See [full bisync description](https://rclone.org/bisync/) for details.
`)

// MakeHelp replaces some dynamic variables for the help docs
func MakeHelp(help string) string {
	replacer := strings.NewReplacer(
		"||", "`",
		"{MAXDELETE}", strconv.Itoa(DefaultMaxDelete),
		"{CHECKFILE}", DefaultCheckFilename,
		"{WORKDIR}", DefaultWorkdir,
	)
	return replacer.Replace(help)
}
