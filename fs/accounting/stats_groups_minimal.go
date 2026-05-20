//go:build minimal

package accounting

import (
	"context"
	"sync"

	"github.com/rclone/rclone/fs"
)

const globalStats = "global_stats"

var groups *statsGroups

func init() {
	// Init stats container
	groups = newStatsGroups()
}

type statsGroupCtx int64

const statsGroupKey statsGroupCtx = 1

// WithStatsGroup returns copy of the parent context with assigned group.
func WithStatsGroup(parent context.Context, group string) context.Context {
	return context.WithValue(parent, statsGroupKey, group)
}

// StatsGroupFromContext returns group from the context if it's available.
// Returns false if group is empty.
func StatsGroupFromContext(ctx context.Context) (string, bool) {
	statsGroup, ok := ctx.Value(statsGroupKey).(string)
	if statsGroup == "" {
		ok = false
	}
	return statsGroup, ok
}

// Stats gets stats by extracting group from context.
func Stats(ctx context.Context) *StatsInfo {
	group, ok := StatsGroupFromContext(ctx)
	if !ok {
		return GlobalStats()
	}
	return StatsGroup(ctx, group)
}

// StatsGroup gets stats by group name.
func StatsGroup(ctx context.Context, group string) *StatsInfo {
	stats := groups.get(group)
	if stats == nil {
		return NewStatsGroup(ctx, group)
	}
	return stats
}

// GlobalStats returns special stats used for global accounting.
func GlobalStats() *StatsInfo {
	return StatsGroup(context.Background(), globalStats)
}

// NewStatsGroup creates new stats under named group.
func NewStatsGroup(ctx context.Context, group string) *StatsInfo {
	stats := NewStats(ctx)
	stats.startAverageLoop()
	stats.group = group
	groups.set(ctx, group, stats)
	return stats
}

// statsGroups holds a synchronized map of stats
type statsGroups struct {
	mu    sync.Mutex
	m     map[string]*StatsInfo
	order []string
}

// newStatsGroups makes a new statsGroups object
func newStatsGroups() *statsGroups {
	return &statsGroups{
		m: make(map[string]*StatsInfo),
	}
}

// set marks the stats as belonging to a group
func (sg *statsGroups) set(ctx context.Context, group string, stats *StatsInfo) {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	ci := fs.GetConfig(ctx)

	// Limit number of groups kept in memory.
	if len(sg.order) >= ci.MaxStatsGroups {
		group := sg.order[0]
		fs.Debugf(nil, "Max number of stats groups reached removing %s", group)
		delete(sg.m, group)
		r := (len(sg.order) - ci.MaxStatsGroups) + 1
		sg.order = sg.order[r:]
	}

	// Exclude global stats from listing
	if group != globalStats {
		sg.order = append(sg.order, group)
	}
	sg.m[group] = stats
}

// get gets the stats for group, or nil if not found
func (sg *statsGroups) get(group string) *StatsInfo {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	stats, ok := sg.m[group]
	if !ok {
		return nil
	}
	return stats
}

func (sg *statsGroups) names() []string {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	return sg.order
}

// sum returns aggregate stats that contains summation of all groups.
func (sg *statsGroups) sum(ctx context.Context) *StatsInfo {
	startTime := GlobalStats().startTime
	sg.mu.Lock()
	defer sg.mu.Unlock()

	sum := NewStats(ctx)
	for _, stats := range sg.m {
		stats.mu.RLock()
		{
			sum.bytes += stats.bytes
			sum.errors += stats.errors
			if sum.lastError == nil && stats.lastError != nil {
				sum.lastError = stats.lastError
			}
			sum.fatalError = sum.fatalError || stats.fatalError
			sum.retryError = sum.retryError || stats.retryError
			if stats.retryAfter.After(sum.retryAfter) {
				// Update the retryAfter field only if it is a later date than the current one in the sum
				sum.retryAfter = stats.retryAfter
			}
			sum.checks += stats.checks
			sum.checking.merge(stats.checking)
			sum.checkQueue += stats.checkQueue
			sum.checkQueueSize += stats.checkQueueSize
			sum.transfers += stats.transfers
			sum.transferring.merge(stats.transferring)
			sum.transferQueue += stats.transferQueue
			sum.transferQueueSize += stats.transferQueueSize
			sum.listed += stats.listed
			sum.renames += stats.renames
			sum.renameQueue += stats.renameQueue
			sum.renameQueueSize += stats.renameQueueSize
			sum.deletes += stats.deletes
			sum.deletesSize += stats.deletesSize
			sum.deletedDirs += stats.deletedDirs
			sum.inProgress.merge(stats.inProgress)
			sum.startedTransfers = append(sum.startedTransfers, stats.startedTransfers...)
			sum.oldTimeRanges = append(sum.oldTimeRanges, stats.oldTimeRanges...)
			sum.oldDuration += stats.oldDuration
			stats.average.mu.Lock()
			sum.average.speed += stats.average.speed
			stats.average.mu.Unlock()
			sum.serverSideCopies += stats.serverSideCopies
			sum.serverSideCopyBytes += stats.serverSideCopyBytes
			sum.serverSideMoves += stats.serverSideMoves
			sum.serverSideMoveBytes += stats.serverSideMoveBytes
		}
		stats.mu.RUnlock()
	}
	sum.startTime = startTime
	return sum
}

func (sg *statsGroups) reset() {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	for _, stats := range sg.m {
		stats.ResetErrors()
		stats.ResetCounters()
	}

	sg.m = make(map[string]*StatsInfo)
	sg.order = nil
}

// delete removes all references to the group.
func (sg *statsGroups) delete(group string) {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	stats := sg.m[group]
	if stats == nil {
		return
	}
	stats.ResetErrors()
	stats.ResetCounters()
	delete(sg.m, group)

	// Remove group reference from the ordering slice.
	tmp := sg.order[:0]
	for _, g := range sg.order {
		if g != group {
			tmp = append(tmp, g)
		}
	}
	sg.order = tmp
}
