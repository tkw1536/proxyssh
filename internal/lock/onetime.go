// Package lock provides various locking utilities.
package lock

import (
	"sync/atomic"
)

// OneTime is an object that can be locked exactly once.
//
// The zero value is ready to use.
// A OneTime must not be copied after creation.
//
// The use-case for this object is an alternative to the sync.Once object.
// sync.Once has two important downsides.
// First any call to .Do() does not return before the action has been performed.
// Second it requires a closure to be passed, resulting in additional code complexity.
// OneTime works around both of these by providing a single Lock() function that returns a boolean.
//
//	 type whatever struct { lock OneTime }
//	 func (w *whatever) DoSomethingOnlyOnce() {
//			if !w.lock.Lock() { // if the action has been started elsewhere, return immediatly.
//				return
//			}
//			// ... action to perform ...
//		}
type OneTime struct {
	noCopy noCopy

	locked uint32 // 1 when locked, 0 when not
}

// Lock attempts to aquire this lock and returns if it was successfull.
//
// Only the first call to this method will return true, all subsequent calls will return false.
func (ol *OneTime) Lock() bool {

	// One could use a mutex and a boolean variable here.
	// However that is slow; instead we can use the atomic package to CompareAndSwap.

	return atomic.CompareAndSwapUint32(&ol.locked, 0, 1)
}

// Do attempts to lock ol and calls f() when successfull.
// Other calls to Do() or Lock() may return before f has returned.
func (ol *OneTime) Do(f func()) bool {
	if !ol.Lock() {
		return false
	}

	f()
	return true
}

// noCopy marks an object to not be copied.
// It is of size 0, and does not contain any data.
// It implements the sync.Locker interface.
//
// See https://stackoverflow.com/a/52495303/11451137.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
