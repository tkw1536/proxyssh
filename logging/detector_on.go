//go:build leak
// +build leak

package logging

// This file enables the memory leak detector.

// MemoryLeakEnabled indicates if the memory leak detector is enabled.
const MemoryLeakEnabled = true

// MemoryLeakDetector represents the either enabled or disabled memory leak detector.
// See MemoryLeakDetectorInterface for documentation.
type MemoryLeakDetector = *MemoryLeakDetectorOn

// NewLeakDetector creates a new MemoryLeakDetector object.
func NewLeakDetector() MemoryLeakDetector {
	return &MemoryLeakDetectorOn{}
}
