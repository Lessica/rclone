//go:build !minimal

package accounting

import (
	"context"
	"errors"
	"fmt"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/rc"
)

// read and set the bandwidth limits
func (tb *tokenBucket) rcBwlimit(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	if in["rate"] != nil {
		bwlimit, err := in.GetString("rate")
		if err != nil {
			return out, err
		}
		var bws fs.BwTimetable
		err = bws.Set(bwlimit)
		if err != nil {
			return out, fmt.Errorf("bad bwlimit: %w", err)
		}
		if len(bws) != 1 {
			return out, errors.New("need exactly 1 bandwidth setting")
		}
		bw := bws[0]
		tb.SetBwLimit(bw.Bandwidth)
	}
	tb.mu.RLock()
	bytesPerSecond := int64(-1)
	if tb.curr[TokenBucketSlotAccounting] != nil {
		bytesPerSecond = int64(tb.curr[TokenBucketSlotAccounting].Limit())
	}
	var bp = fs.BwPair{Tx: -1, Rx: -1}
	if tb.curr[TokenBucketSlotTransportTx] != nil {
		bp.Tx = fs.SizeSuffix(tb.curr[TokenBucketSlotTransportTx].Limit())
	}
	if tb.curr[TokenBucketSlotTransportRx] != nil {
		bp.Rx = fs.SizeSuffix(tb.curr[TokenBucketSlotTransportRx].Limit())
	}
	tb.mu.RUnlock()
	out = rc.Params{
		"rate":             bp.String(),
		"bytesPerSecond":   bytesPerSecond,
		"bytesPerSecondTx": int64(bp.Tx),
		"bytesPerSecondRx": int64(bp.Rx),
	}
	return out, nil
}

// Remote control for the token bucket
func init() {
	rc.Add(rc.Call{
		Path:  "core/bwlimit",
		Fn:    TokenBucket.rcBwlimit,
		Title: "Set the bandwidth limit.",
		Help: `
This sets the bandwidth limit to the string passed in. This should be
a single bandwidth limit entry or a pair of upload:download bandwidth.

Eg

    rclone rc core/bwlimit rate=off
    {
        "bytesPerSecond": -1,
        "bytesPerSecondTx": -1,
        "bytesPerSecondRx": -1,
        "rate": "off"
    }
    rclone rc core/bwlimit rate=1M
    {
        "bytesPerSecond": 1048576,
        "bytesPerSecondTx": 1048576,
        "bytesPerSecondRx": 1048576,
        "rate": "1M"
    }
    rclone rc core/bwlimit rate=1M:100k
    {
        "bytesPerSecond": 1048576,
        "bytesPerSecondTx": 1048576,
        "bytesPerSecondRx": 131072,
        "rate": "1M"
    }


If the rate parameter is not supplied then the bandwidth is queried

    rclone rc core/bwlimit
    {
        "bytesPerSecond": 1048576,
        "bytesPerSecondTx": 1048576,
        "bytesPerSecondRx": 1048576,
        "rate": "1M"
    }

The format of the parameter is exactly the same as passed to --bwlimit
except only one bandwidth may be specified.

In either case "rate" is returned as a human-readable string, and
"bytesPerSecond" is returned as a number.
`,
	})
}
