package logging

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tkw1536/proxyssh/internal/lock"
)

// MemoryLeakDetectorInterface represents a Memory Leak Detector.
// It is used to ensure that no concurrent processes have caused a memory leak within sessions.
//
// This interface is not used at runtime, it serves only for documentation purposes.
// The MemoryLeakDetectorInterface can be turned on or off at build time using the "leak" build tag.
// This sets the MemoryLeakDetector type alias is set either to MemoryLeakDetectorOn or MemoryLeakDetectorOff.
// These contain the actual implementation.
//
// At runtime, all instances should be created using NewLeakDetector and never MemoryLeakDetectorOn or MemoryLeakDetectorOff.
// They are still exposed, so that both implementations can be tested.
//
// When the MemoryLeakDetectorInterface is disabled, all functions in this interface are noops.
// As opposed to using this interface, that enables the disabled variant to be compiled away.
type MemoryLeakDetectorInterface interface {
	// Add and Done are used to register async processes with this LeakDetector.
	//
	// A call to Add() should be performed when an async action is started.
	// A call to Done() with the same name should be performed when the async action is completed as expected.
	//
	// Names should be unique. Upon reuse, Add() and Done() call panic().
	Add(name string)
	Done(name string)

	// Finish checks if all calls to Add() have been undone by a call to Done().
	// It should be called syncronously.
	//
	// If this is not the case within MemoryLeakTimeout, it prints to logger an appropriate error message.
	// If this is the case, it prints a short message confirming that the leak check completed.
	Finish(logger Logger, s LogSessionOrContext)
}

// the implementation of MemoryLeakDetector fullfills the LeakDetector interface
func init() {
	// MemoryLeakDetector alias implements MemoryLeakDetectorInterface
	var impl MemoryLeakDetector
	var _ MemoryLeakDetectorInterface = impl

	// Enabled and disabled variants implement the leak detector
	var _ MemoryLeakDetectorInterface = (*MemoryLeakDetectorOn)(nil)
	var _ MemoryLeakDetectorInterface = MemoryLeakDetectorOff{}
}

// MemoryLeakTimeout is the default used by the memory detector to trigger.
const MemoryLeakTimeout time.Duration = time.Second

// leakDetectorState containts a global leak detector state.
// It should be accessed using sync/atomic calls only.
var leakDetectorState struct {
	Success, Failure uint64 // multiple counter

	workers lock.WorkGroup // ongoing calls to any leak detector
}

// ResetGlobalLeakDetectorStats resets the stats tracker used for the global leak detector.
func ResetGlobalLeakDetectorStats() {
	leakDetectorState.workers.Lock()
	defer leakDetectorState.workers.Unlock()

	leakDetectorState.workers.Wait()

	atomic.StoreUint64(&leakDetectorState.Success, 0)
	atomic.StoreUint64(&leakDetectorState.Failure, 0)
}

// GetGlobalLeakDetectorStats gets statistics about global leak detector calls.
// Blocks until a leak detector is available.
func GetGlobalLeakDetectorStats() (success, failure uint64) {
	leakDetectorState.workers.Lock()
	defer leakDetectorState.workers.Unlock()

	leakDetectorState.workers.Wait()

	return atomic.LoadUint64(&leakDetectorState.Success), atomic.LoadUint64(&leakDetectorState.Failure)
}

//
//  Enabled Memory Leak Detector
//

// MemoryLeakDetectorOn is the enabled variant of MemoryLeakDetectorInterface.
type MemoryLeakDetectorOn struct {
	wg sync.WaitGroup
	m  sync.Map
}

// Add implements MemoryLeakDetectorInterface.Add.
func (d *MemoryLeakDetectorOn) Add(name string) {
	d.wg.Add(1) // tell the workgroup that we had one item!

	// grab the function name of the caller
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])

	// store it it in the map
	_, loaded := d.m.LoadOrStore(name, fmt.Sprintf("%q, near %s:%d", f.Name(), file, line))
	if loaded {
		panic("LeakDetectorOn: Add() reused key")
	}
}

// Done implements MemoryLeakDetectorInterface.Done.
func (d *MemoryLeakDetectorOn) Done(name string) {
	d.wg.Done()
	if _, ok := d.m.Load(name); !ok {
		panic("LeakDetectorOn: Done() removed unknown key")
	}
	d.m.Delete(name)
}

// Finish implements MemoryLeakDetectorInterface.Finish.
func (d *MemoryLeakDetectorOn) Finish(logger Logger, s LogSessionOrContext) {
	// inform the global state that we are adding more work!
	leakDetectorState.workers.Add(1)

	go func() {
		defer leakDetectorState.workers.Done()

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
	}()
}

func (d *MemoryLeakDetectorOn) unfire(logger Logger, s LogSessionOrContext) {
	atomic.AddUint64(&leakDetectorState.Success, 1)
	FmtSSHLog(logger, s, "leak_ok")
}

func (d *MemoryLeakDetectorOn) fire(logger Logger, s LogSessionOrContext) {
	atomic.AddUint64(&leakDetectorState.Failure, 1)
	FmtSSHLog(logger, s, "leak_fail")
	d.m.Range(func(k, v interface{}) bool {
		logger.Printf("Leak Detector: %s (%q)", v, k)
		return true
	})
}

//
//  Disabled Memory Leak Detector
//

// MemoryLeakDetectorOff is the disabled variant of MemoryLeakDetectorInterface.
type MemoryLeakDetectorOff struct{}

// Add implements MemoryLeakDetectorInterface.Add.
func (d MemoryLeakDetectorOff) Add(name string) {}

// Done implements MemoryLeakDetectorInterface.Done.
func (d MemoryLeakDetectorOff) Done(name string) {}

// Finish implements MemoryLeakDetectorInterface.Finish.
func (d MemoryLeakDetectorOff) Finish(logger Logger, s LogSessionOrContext) {}
