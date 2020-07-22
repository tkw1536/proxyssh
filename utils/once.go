package utils

import (
	"sync"
)

// Once is an object that can perform any action once
type Once struct {
	mutex     sync.Mutex
	performed bool
}

// Once performs the action it is passed the first time it is called
// and otherwise does nothing
func (o *Once) Once(action func()) bool {
	if !o.shouldPerform() {
		return false
	}

	action()
	return true
}

// shouldPerform returns if an action should be performed or not
// when returning true, expects that the caller will perform the action
// when returning false, an action has already been performed
func (o *Once) shouldPerform() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.performed {
		return false
	}

	o.performed = true
	return true
}
