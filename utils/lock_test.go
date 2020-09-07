package utils

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestOneTime_Lock(t *testing.T) {
	onetime := &OneTime{} // object being tested

	var trues uint64
	var falses uint64

	// create a waitgroup
	N := 100000
	var wg sync.WaitGroup
	wg.Add(N)

	// count trues and falses returned by Lock
	for i := 0; i < N; i++ {
		go func() {
			if onetime.Lock() {
				atomic.AddUint64(&trues, 1)
			} else {
				atomic.AddUint64(&falses, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	wantTrues := uint64(1)
	if trues != wantTrues {
		t.Errorf("OneTime.Lock() returned %d true(s), expected %d. ", trues, wantTrues)
	}

	wantFalses := uint64(N - 1)
	if falses != wantFalses {
		t.Errorf("OneTime.Lock() returned %d false(s), expected %d. ", falses, wantFalses)
	}
}

func BenchmarkOneTime_Lock(b *testing.B) {

	onetime := &OneTime{} // object being benchmarked

	var trues uint64
	var falses uint64

	// create a waitgroup
	var wg sync.WaitGroup
	wg.Add(b.N)

	// count trues and falses returned by Lock
	for i := 0; i < b.N; i++ {
		go func() {
			if onetime.Lock() {
				atomic.AddUint64(&trues, 1)
			} else {
				atomic.AddUint64(&falses, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()

}
