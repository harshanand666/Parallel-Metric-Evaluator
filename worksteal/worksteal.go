package worksteal

import (
	"sync/atomic"
	"unsafe"
)

// Represents a single row of input data
type Task struct {
	Record *[]string
	next   *Task
	prev   *Task
}

// Unbounded deque for each thread
type Deque struct {
	head      unsafe.Pointer
	tail      unsafe.Pointer
	ThreadNum int
}

// Returns a pointer to a new deque after populating it with tasks
func NewDeque(records *[][]string, start int, end int, idx int) *Deque {
	node := &Task{next: nil, prev: nil}
	curQueue := Deque{head: unsafe.Pointer(node), tail: unsafe.Pointer(node), ThreadNum: idx}
	for i := start; i < end; i++ {
		curQueue.PushBottom(&Task{Record: &((*records)[i])})
	}
	return &curQueue
}

// Adds new task to the deque
func (d *Deque) PushBottom(t *Task) {
	(*Task)(d.tail).next = t
	t.prev = (*Task)(d.tail)
	d.tail = unsafe.Pointer(t)
}

// Called by a stealer to take a task from the top of the deque
func (d *Deque) PopTop() *Task {

	head := atomic.LoadPointer(&(d.head))
	tail := atomic.LoadPointer(&(d.tail))

	if head == tail {
		// no tasks remaining
		return nil
	}
	stolenTask := (*Task)(head).next
	if atomic.CompareAndSwapPointer((&d.head), head, unsafe.Pointer(stolenTask)) {
		// If CAS succeeds, task stolen
		return stolenTask
	}
	return nil
}

// Called by own thread to take a task from the bottom of the deque
func (d *Deque) PopBottom() *Task {

	head := atomic.LoadPointer(&(d.head))
	tail := atomic.LoadPointer(&(d.tail))

	if head == tail {
		// No elements in deque
		return nil
	}
	curTask := (*Task)(tail)
	if (*Task)(head).next != (*Task)(tail) {
		// More than one element, take task and update tail
		atomic.StorePointer(&d.tail, unsafe.Pointer(curTask.prev))
		return curTask
	}
	if (*Task)(head).next == (*Task)(tail) {
		// Only one element in the queue, need to use atomic CAS to take task
		if atomic.CompareAndSwapPointer((&d.head), head, unsafe.Pointer(curTask)) {
			return curTask
		}
	}

	return nil
}
