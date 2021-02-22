package asyncio

import "sync"

// size used for copy buffers
const copyBufferSize = 32 * 1024

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, copyBufferSize)
	},
}

// bufferFromPool returns a new []byte with length size.
// It should be returned to the pool using bufferPool.Put(s)
func bufferFromPool(l int) []byte {
	s := bufferPool.Get().([]byte)
	if cap(s) < l { // too small?
		bufferPool.Put(s)
		return make([]byte, l)
	}
	return s[:l]
}
