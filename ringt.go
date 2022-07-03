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

// RingT is a generic ring buffer that holds pointers to a generic type T.
// When popping an item from the tail of the ring, we set it's pointer to nil,
// to ensure that the garbage collector can reclaim the memory for that item.
type RingT[T any] struct {
	items   []*T // len(items) is a power of 2.
	mask    uint // mask = len(items) - 1
	tail    uint // read from tail
	head    uint // write into head
	maxSize int
}

// NewRingT creates a new ring buffer with the specified maximum size.
// The maximum size must be at least 1.
// The ring's underlying buffer is grown incrementally in powers of 2,
// so the maxSize is not allocated up front.
func NewRingT[T any](maxSize int) RingT[T] {
	if maxSize < 1 {
		panic("RingT size must be at least 1")
	}
	return RingT[T]{
		maxSize: maxSize,
	}
}

// MaxSize is the maximum number of elements in the ring buffer
func (r *RingT[T]) MaxSize() int {
	return r.maxSize
}

// IsFull returns true if the ring buffer is full, and adding
// another item will cause the oldest item to be popped.
func (r *RingT[T]) IsFull() bool {
	return r.Len() == r.maxSize
}

// Len returns the number of elements in the buffer
func (r *RingT[T]) Len() int {
	return int((r.head - r.tail) & r.mask)
}

// Next returns the next item in the ring, or nil if the ring is empty
func (r *RingT[T]) Next() *T {
	if r.Len() == 0 {
		return nil
	}
	t := r.tail
	r.tail = (r.tail + 1) & r.mask
	item := r.items[t]
	r.items[t] = nil // erase item, so that the garbage collector can do it's job
	return item
}

// Peek returns the Tail+i element from the buffer.
// Peek(0) returns the same result as Next(), except that Peek() does not
// change any state.
func (r *RingT[T]) Peek(i int) *T {
	length := (r.head - r.tail) & r.mask
	ui := uint(i)
	if ui >= length {
		return nil
	}
	j := (r.tail + ui) & r.mask
	return r.items[j]
}

// Add an item to the buffer.
// If the buffer is full, erase the oldest item.
func (r *RingT[T]) Add(item *T) {
	if r.Len() == r.maxSize {
		// erase oldest item
		r.Next()
	}

	if len(r.items) == 0 || r.Len() == len(r.items)-1 {
		// need to grow array
		newSize := len(r.items) * 2
		if newSize < 2 {
			newSize = 2
		}
		newItems := make([]*T, newSize, newSize)
		n := r.Len()
		for i := 0; i < n; i++ {
			item := r.Next()
			newItems[i] = item
		}
		r.items = newItems
		r.mask = uint(newSize) - 1
		r.tail = 0
		r.head = uint(n)
	}

	r.items[r.head] = item
	r.head = (r.head + 1) & r.mask
}
