//go:build !leak
// +build !leak

package logging

// This file disables the memory leak detector.

// MemoryLeakEnabled indicates if the memory leak detector is enabled.
// See LeakDetector interface.
const MemoryLeakEnabled = false

// MemoryLeakDetector represents the either enabled or disabled memory leak detector.
// See MemoryLeakDetectorInterface for documentation.
type MemoryLeakDetector = MemoryLeakDetectorOff

// NewLeakDetector creates a new MemoryLeakDetector object.
func NewLeakDetector() MemoryLeakDetector {
	return MemoryLeakDetectorOff{}
}
