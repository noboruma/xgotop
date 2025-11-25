//go:build linux
// +build linux

package main

import (
	"golang.org/x/sys/unix"
)

// getMonotonicNs returns monotonic time in nanoseconds, matching bpf_ktime_get_ns()
func getMonotonicNs() uint64 {
	var ts unix.Timespec
	// CLOCK_MONOTONIC matches what bpf_ktime_get_ns uses in the kernel
	err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &ts)
	if err != nil {
		// This shouldn't fail on Linux, but if it does, panic
		panic("failed to get monotonic clock: " + err.Error())
	}
	return uint64(ts.Sec*1e9 + ts.Nsec)
}