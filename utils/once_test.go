package utils

import (
	"sync"
	"testing"
)

func Test_Once_Once(t *testing.T) {
	t.Run("single action is only performed once", func(t *testing.T) {
		var actionPerformed int
		var o Once
		action := func() {
			actionPerformed = actionPerformed + 1
		}

		// perform it the first time
		firstTime := o.Once(action)
		if actionPerformed != 1 {
			t.Errorf("Once.Once() did not perform the action once")
			return
		}
		if firstTime != true {
			t.Errorf("Once.Once() = %v, expected %v", firstTime, true)
		}

		// don't perform it the second time
		secondTime := o.Once(action)
		if actionPerformed != 1 {
			t.Errorf("Once.Once() performed action a second time")
		}
		if secondTime != false {
			t.Errorf("Once.Once() = %v, expected %v", firstTime, false)
		}
	})

	t.Run("multiple actions are only performed once", func(t *testing.T) {
		var actionPerformed int
		var o Once
		firstAction := func() {
			actionPerformed = actionPerformed + 1
		}
		secondAction := func() {
			actionPerformed = actionPerformed + 2
		}

		// perform it the first time
		firstTime := o.Once(firstAction)
		if actionPerformed != 1 {
			t.Errorf("Once.Once() did not perform the action once")
			return
		}
		if firstTime != true {
			t.Errorf("Once.Once() = %v, expected %v", firstTime, true)
		}

		// don't perform it the Nsecond time
		secondTime := o.Once(secondAction)
		if actionPerformed != 1 {
			t.Errorf("Once.Once() performed action a second time")
		}
		if secondTime != false {
			t.Errorf("Once.Once() = %v, expected %v", firstTime, false)
		}
	})
}
func Benchmark_Test_Once_Once(b *testing.B) {
	// setup action and once object
	var actionPerformed int
	var o Once

	// create a list of results
	// and a wait group
	results := make([]bool, b.N)
	var w sync.WaitGroup
	w.Add(b.N)

	// run all of them in paralell and count the results
	for i := 0; i < b.N; i++ {
		go func(i int) {
			results[i] = o.Once(func() {
				actionPerformed = actionPerformed + 1
			})
			w.Done()
		}(i)
	}
	w.Wait()

	// the action was performed once
	if actionPerformed != 1 {
		b.Errorf("Once.Once() did not perform the action")
		return
	}

	// ensure that true is only returned once
	var trueC bool
	for _, r := range results {
		if r {
			if trueC {
				b.Errorf("Once.Once() returned true more than once")
				return
			}
			trueC = true
		}
	}
}
