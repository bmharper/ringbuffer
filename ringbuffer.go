/*
Package ringbuffer implements a byte-based ring buffer and a generic item ring buffer.

The ring buffer has the following properties:
	* Total size is always a power of 2
	* Capacity of the buffer is total size - 1
	* Head == Tail is an empty buffer
	* The buffer is full when Head is one less than Tail
	* The buffer grows itself in powers of 2

Thanks to https://fgiesen.wordpress.com/2010/12/14/ring-buffers-and-queues/ for explaining how to implement a ring buffer correctly.

*/
package ringbuffer

import (
	"io"
)

// Start buffer at 64 bytes. This just seems like a reasonable minimum.
const DefaultSize = 64

// The zero value for Ring is an empty buffer ready to use.
type Ring struct {
	head uint
	tail uint
	data []byte
}

// Return the number of unread bytes in the buffer
func (r *Ring) Len() int {
	return int((r.head - r.tail) & r.mask())
}

// Grow the buffer sufficiently so that you can write numBytes into it.
// The returned slice is not guaranteed to be large enough to hold numBytes. If
// the slice is not large enough, then it means that the requested range falls off the edge
// of the circular buffer, and you'll need to call this twice.
// Note that the slice returned points directly into the ring buffer.
// After calling this function, the Len() will be increased by the length of the returned slice,
// and the head of the buffer will point to the end of the slice.
// This function exists because it makes it possible, in certain cases, to get away with fewer memory copies
// than if you were to use the Write() interface.
func (r *Ring) DirectWrite(numBytes int) []byte {
	r.ensureCapacity(uint(r.Len() + numBytes))
	if int(r.end()-r.head) < numBytes {
		numBytes = int(r.end() - r.head)
	}
	slice := r.data[int(r.head) : int(r.head)+numBytes]
	r.head = (r.head + uint(numBytes)) & r.mask()
	return slice
}

// Reads the requested number of bytes, but possibly returns less than
// the requested number. If Tail + numBytes wraps around, then the returned
// slice will only contain bytes Capacity - Tail bytes. You need to perform
// a subsequent read to read the remaining portion of the buffer.
// Note that the slice returned points directly into the ring buffer.
// This function exists because it makes it possible, in certain cases, to get away with fewer memory copies
// than if you were to use the Read() interface.
func (r *Ring) DirectRead(numBytes int) []byte {
	if numBytes > r.Len() {
		numBytes = r.Len()
	}
	if numBytes > int(r.end()-r.tail) {
		numBytes = int(r.end() - r.tail)
	}
	if numBytes <= 0 {
		return nil
	}
	res := r.data[r.tail : r.tail+uint(numBytes)]
	r.tail = (r.tail + uint(numBytes)) & r.mask()
	return res
}

// Implements io.Reader
func (r *Ring) Read(b []byte) (int, error) {
	s1 := r.DirectRead(len(b))
	copy(b, s1)
	s2 := r.DirectRead(len(b) - len(s1))
	copy(b[len(s1):], s2)

	total := len(s1) + len(s2)
	if total == 0 && r.Len() == 0 {
		return total, io.EOF
	} else {
		return total, nil
	}
}

// Implements io.Writer
func (r *Ring) Write(b []byte) (int, error) {
	b1 := r.DirectWrite(len(b))
	copy(b1, b)
	if len(b1) != len(b) {
		b2 := r.DirectWrite(len(b) - len(b1))
		copy(b2, b[len(b1):])
	}
	return len(b), nil
}

// End of the buffer
func (r *Ring) end() uint {
	return uint(len(r.data))
}

func (r *Ring) mask() uint {
	return uint(len(r.data)) - 1
}

// Ensure our capacity is large enough to hold forBytes bytes. Grow by powers of 2.
func (r *Ring) ensureCapacity(forBytes uint) {
	// The +1 here is because we can only store len(r.data)-1 objects.
	needCap := forBytes + uint(r.Len()) + 1
	if needCap <= uint(len(r.data)) {
		return
	}
	orgCap := uint(len(r.data))
	cap := orgCap
	if cap < DefaultSize {
		cap = DefaultSize
	}
	for cap < needCap {
		cap *= 2
	}
	extra := int(cap - orgCap)
	for i := 0; i < extra; i++ {
		r.data = append(r.data, 0)
	}
	if r.head < r.tail {
		// Handle the scenario where the head is behind the tail (numerically)
		// [  H T  ]   =>  [    T     H    ]
		// [c - a b]   =>  [- - a b c - - -]
		buf := r.data
		copy(buf[orgCap:orgCap+r.head], buf[0:r.head])
		r.head += orgCap

	}
}
