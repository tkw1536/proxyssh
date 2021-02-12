package lock

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func ExampleWorkGroup() {
	wg := &WorkGroup{}

	// perform some work
	wg.Add(1)
	go func() {
		defer wg.Done()

		fmt.Println("... first set of work ...")
	}()

	// lock the workgroup
	wg.Lock()

	// schedule some work (which won't happen until Unlock() has been called.
	workSched := make(chan struct{}) // closed once the work below has started
	go func() {
		wg.Add(1)
		close(workSched) // work has been scheduled!

		defer wg.Done()

		fmt.Println("... second set of work ...")
	}()

	wg.Wait()
	fmt.Println("first set of work done")
	wg.Unlock()

	<-workSched // wait until the work has been scheduled!

	wg.Wait()
	fmt.Println("second set of work done")

	// Output: ... first set of work ...
	// first set of work done
	// ... second set of work ...
	// second set of work done
}

func TestWorkGroup(t *testing.T) {
	N := 10000
	N64 := int64(N)

	// create a waitgroup
	wg := WorkGroup{}
	var counter int64

	// add 1 to the counter N times
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	// lock for further adds
	wg.Lock()

	// do the same thing again
	// but because of Lock(), this should block.
	schedChan := make(chan struct{})
	go func() {
		wg.Add(N)
		close(schedChan)
		for i := 0; i < N; i++ {
			go func() {
				defer wg.Done()
				atomic.AddInt64(&counter, 1)
			}()
		}
	}()

	// wait and check that we have a count of N
	wg.Wait()
	if counter != N64 {
		t.Fatalf("got counter = %d, expected %d", counter, N64)
	}

	// reset the counter and unlock!
	counter = 0
	wg.Unlock()

	// wait for the next group of goroutines to finish
	<-schedChan // scheduling done!
	wg.Wait()
	if counter != N64 {
		t.Fatalf("got counter = %d, expected %d", counter, N64)
	}

}
