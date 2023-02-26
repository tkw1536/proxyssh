package lock

import "sync"

// WorkGroup is a combination of an sync.Mutex and sync.WaitGroup.
//
// It provides both a sync.Locker and a Add() and Done() methods.
// When locking the locker, all calls to Add() will block until it is unlocked.
//
// The zero value is ready to use.
type WorkGroup struct {
	mutex     sync.RWMutex
	waitgroup sync.WaitGroup
}

// Lock locks all WorkGroup calls and blocks future Add() calls until Unlock() is called.
// Only a single Lock() call is possible at the same time.
func (wg *WorkGroup) Lock() {
	wg.mutex.Lock()
}

// Unlock undoes a call to Lock()
func (wg *WorkGroup) Unlock() {
	wg.mutex.Unlock()
}

// Add blocks until all calls to Lock() have been undone and then adds n to the underlying Waitgroup.
func (wg *WorkGroup) Add(n int) {
	wg.mutex.RLock()
	defer wg.mutex.RUnlock()

	wg.waitgroup.Add(n)
}

// Done calls Done() on the underlying WaitGroup
func (wg *WorkGroup) Done() {
	wg.waitgroup.Done()
}

// Wait calls Wait() on the underlying WaitGroup.
// A typical use of this function is like:
//
//	wg.Lock()
//	defer wg.Unlock()
//	wg.Wait()
//
//	// perform some work that during which no new jobs are added.
func (wg *WorkGroup) Wait() {
	wg.waitgroup.Wait()
}
