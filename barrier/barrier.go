package barrier

import "sync"

// Custom Implementation of a barrier
type Barrier struct {
	cond       *sync.Cond
	Ctr        int
	TotalCount int
}

// Calling thread suspends till all threads have arrived
func (b *Barrier) Await() {
	b.cond.L.Lock()
	b.Ctr += 1
	if b.Ctr < b.TotalCount {
		b.cond.Wait()
	} else {
		b.cond.Broadcast()
	}
	b.cond.L.Unlock()
}

// Returns a pointer to a new barrier object
func NewBarrier() *Barrier {
	mutex := &sync.Mutex{}
	cond := sync.NewCond(mutex)
	return &Barrier{cond: cond, Ctr: 0, TotalCount: 0}
}
