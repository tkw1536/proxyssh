package utils

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
//
//  type whatever struct { lock OneTime }
//  func (w *whatever) DoSomethingOnlyOnce() {
//		if !w.lock.Lock() { // if the action has been started elsewhere, return immediatly.
//			return
//		}
//		// ... action to perform ...
//	}
//
type OneTime struct {
	noCopy noCopy

	state uint32
}

// Lock attempts to aquire this lock and returns if it was successfull.
//
// Only the first call to this method will return true, all subsequent calls will return false.
func (ol *OneTime) Lock() bool {

	// One could use a mutex and a boolean variable here.
	// However that is slow; instead we can use the atomic package to CompareAndSwap.
	//
	// Because this is a very special low-level operation, we are explicitly ignoring the atomic package warning
	// to not use it.

	// At init time ol.state will be 0.
	// A locked state is indicated by a 1.
	// atom.CompareAndSwap performs the swapping process
	// and also checks if the lock was open before.
	//
	// Since most of the time we have already locked the state
	// (and thus only need to return false) this turns out to be
	// the fastest.
	//
	// This has the side-effect that the call can be efficiently inlined.

	return atomic.CompareAndSwapUint32(&ol.state, 0, 1)
}

// noCopy marks an object to not be copied.
// It is of size 0, and does not contain any data.
// It implements the sync.Locker interface.
//
// See https://stackoverflow.com/a/52495303/11451137.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
