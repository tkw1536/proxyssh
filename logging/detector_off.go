// +build !leak

package logging

import (
	"time"
)

// This file contains code for a deactivated leak detector.
// It contains only no-op functions.
//
// The Detector object intentionally does not have a pointer receiver.
// This is so that calls (in particular the disabled version) can be inlined and compiled away.

// MemoryLeakEnabled indicates if the memory leak detector is enabled.
const MemoryLeakEnabled = false

// MemoryLeakTimeout is the timout used by the memory detector to trigger.
const MemoryLeakTimeout time.Duration = 0

// NewMemoryLeakDetector creates a new Detector object.
func NewMemoryLeakDetector() MemoryLeakDetector {
	return MemoryLeakDetector{}
}

// MemoryLeakDetector is an object that can keep track of async leaks.
// It can be disabled using a build tag.
type MemoryLeakDetector struct{}

// Add indicates that an action should be completed asyncronously.
// Any call to Add() should be undone with a call to Done().
func (d MemoryLeakDetector) Add(name string) {}

// Done indicates that the provided call to Add() has completed.
func (d MemoryLeakDetector) Done(name string) {}

// Finish checks if all calls to Add() have been undone by a call to Done().
// It then prints an appropriate log message.
//
// When timeout is 0, picks a reasonable detault duration.
func (d MemoryLeakDetector) Finish(logger Logger, s SSHSessionOrContext, timeout time.Duration) {}
