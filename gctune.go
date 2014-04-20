// Package gctune allows for easier management of the Go GC behavior.
package gctune

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

var (
	tSize int64 = -1
	mu    sync.RWMutex
)

func init() {
	go monitor()
}

// SetTargetSysSize changes the desired maximum system memory to be used by
// this process. It returns the previously set value, or -1 if inset.
// (This doesn't work yet.)
func SetTargetSysSize(t int64) int64 {
	mu.Lock()
	defer mu.Unlock()
	was := tSize
	tSize = t
	return was
}

func monitor() {
	c := time.Tick(1 * time.Second)
	mem := new(runtime.MemStats)
	origPct := debug.SetGCPercent(100)
	debug.SetGCPercent(origPct)
	for _ = range c {
		runtime.ReadMemStats(mem)
		mu.Lock()
		defer mu.Unlock()
		if tSize < 0 {
			continue
		}
		// Occupancy fraction: 70%. Don't GC before hitting this.
		softLimit := float64(tSize) * 0.7
		pct := softLimit / float64(mem.Alloc) * 100
		fmt.Printf("gctune: pct: %0.5f, target: %d, softLimit: %0.2f, Alloc: %d, Sys: %d\n", pct, tSize, softLimit, mem.Alloc, mem.Sys)
		if pct < 50 {
			// If this is too low, GC frequency increases too much.
			pct = 50
		}
		debug.SetGCPercent(int(pct))
		if mem.Sys > uint64(tSize*70/100) {
			fmt.Println("freeing")
			debug.FreeOSMemory()
		}
	}
}

// debug.FreeOSMemory
// debug.SetGCPercent
