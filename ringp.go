package ringbuffer

// Example
//
// length: 8
// tail:   1
// head:   3
// number of elements in buffer: 2
//
// 0      1      2      3      4      5      6      7
//      tail          head

// RingP is a generic ring buffer that holds pointers to a generic type T.
// This doesn't hold pointers of T, but concrete instances of them.
// Also, the ring has a static buffer - it is allocated at creation,
// and never changes thereafter.
type RingP[T any] struct {
	items []T  // len(items) is a power of 2.
	mask  uint // mask = len(items) - 1
	tail  uint // read from tail
	head  uint // write into head
}

// NewRingP creates a new ring buffer with the specified maximum size.
// sizePlus1 must be a power of 2.
// The maximum number of elements in the ring is sizePlus1 - 1
func NewRingP[T any](sizePlus1 int) RingP[T] {
	if (sizePlus1&(sizePlus1-1)) != 0 || sizePlus1 < 2 {
		panic("sizePlus1 must be a power of 2, and minimum 2")
	}
	return RingP[T]{
		items: make([]T, sizePlus1),
		mask:  uint(sizePlus1) - 1,
		tail:  0,
		head:  0,
	}
}

// Capacity is the capacity of the ring buffer, which is 2^N - 1
func (r *RingP[T]) Capacity() int {
	return int(r.mask)
}

// IsFull returns true if the ring buffer is full, and adding
// another item will cause the oldest item to be popped.
func (r *RingP[T]) IsFull() bool {
	return r.Len() == int(r.mask)
}

// Len returns the number of elements in the buffer
func (r *RingP[T]) Len() int {
	return int((r.head - r.tail) & r.mask)
}

// Next returns the next item in the ring, or the zero object if the ring is empty
func (r *RingP[T]) Next() T {
	if r.Len() == 0 {
		var zero T
		return zero
	}
	t := r.tail
	r.tail = (r.tail + 1) & r.mask
	item := r.items[t]
	return item
}

// Peek returns the Tail+i element from the buffer.
// Peek(0) returns the same result as Next(), except that Peek() does not
// change any state.
func (r *RingP[T]) Peek(i int) T {
	length := (r.head - r.tail) & r.mask
	ui := uint(i)
	if ui >= length {
		var zero T
		return zero
	}
	j := (r.tail + ui) & r.mask
	return r.items[j]
}

// Add an item to the buffer.
// If the buffer is full, erase the oldest item.
func (r *RingP[T]) Add(item T) {
	if r.IsFull() {
		// erase oldest item
		r.Next()
	}
	r.items[r.head] = item
	r.head = (r.head + 1) & r.mask
}
