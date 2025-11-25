//go:build !linux
// +build !linux

package main

// getMonotonicNs returns monotonic time in nanoseconds
// On non-Linux platforms, we use the Go runtime's monotonic clock
func getMonotonicNs() uint64 {
	panic("unimplemented")
}
