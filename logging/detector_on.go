// +build leak

package logging

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryLeakEnabled indicates if the memory leak detector is enabled.
const MemoryLeakEnabled = true

// MemoryLeakTimeout is the timout used by the memory detector to trigger.
const MemoryLeakTimeout time.Duration = time.Second

// NewMemoryLeakDetector creates a new Detector object.
func NewMemoryLeakDetector() MemoryLeakDetector {
	return MemoryLeakDetector{
		wg: &sync.WaitGroup{},
		m:  &sync.Map{},
	}
}

// The MemoryLeakDetector object intentionally does not have a pointer receiver.
// This is so that calls (in particular the disabled version) can be inlined and compiled away.

// MemoryLeakDetector is an object that can keep track of async leaks.
// It can be disabled using a build tag.
type MemoryLeakDetector struct {
	wg *sync.WaitGroup
	m  *sync.Map
}

// Add indicates that an action should be completed asyncronously.
// Any call to Add() should be undone with a call to Done().
func (d MemoryLeakDetector) Add(name string) {
	d.wg.Add(1) // tell the workgroup that we had one item!

	// grab the function name of the caller
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])

	// store it it in the map
	_, loaded := d.m.LoadOrStore(name, fmt.Sprintf("%q, near %s:%d", f.Name(), file, line))
	if loaded {
		panic("MemoryLeakDetector: Add() reused key")
	}
}

// Done indicates that the provided call to Add() has completed.
func (d MemoryLeakDetector) Done(name string) {
	d.wg.Done()
	if _, ok := d.m.Load(name); !ok {
		panic("MemoryLeakDetector: Done() removed unknown key")
	}
	d.m.Delete(name)
}

// Finish checks if all calls to Add() have been undone by a call to Done().
// It then prints an appropriate log message.
//
// When timeout is 0, picks a reasonable detault duration.
func (d MemoryLeakDetector) Finish(logger Logger, s SSHSessionOrContext, timeout time.Duration) {

	if timeout == 0 {
		timeout = MemoryLeakTimeout
	}

	waiter := make(chan struct{})
	go func() {
		defer close(waiter)
		d.wg.Wait()
	}()

	select {
	case <-waiter: /* everything ok */
		d.unfire(logger, s)
	case <-time.After(MemoryLeakTimeout): /* timeout fired, group didn't exit */
		d.fire(logger, s)
	}
}

func (d MemoryLeakDetector) unfire(logger Logger, s SSHSessionOrContext) {
	FmtSSHLog(logger, s, "leak_ok")
}

func (d MemoryLeakDetector) fire(logger Logger, s SSHSessionOrContext) {
	FmtSSHLog(logger, s, "leak_fail")
	d.m.Range(func(k, v interface{}) bool {
		logger.Printf("Leak Detector: %s (%q)", v, k)
		return true
	})
}
