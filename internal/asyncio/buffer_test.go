package asyncio

import (
	"testing"
)

func Test_bufferFromPool(t *testing.T) {
	tests := []int{0, 10, 100, 1000, 10000}
	for _, s := range tests {
		buffer := bufferFromPool(s)
		if len(buffer) != s {
			t.Errorf("bufferFromPool() returned wrong size")
		}
	}
}
