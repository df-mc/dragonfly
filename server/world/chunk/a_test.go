package chunk

import (
	"sync"
	"testing"
)

func BenchmarkMutex(b *testing.B) {
	a := new(sync.Mutex)
	for i := 0; i < b.N; i++ {
		a.Lock()
		a.Unlock()
	}
}
