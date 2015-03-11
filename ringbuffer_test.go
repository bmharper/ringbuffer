package ringbuffer

import (
	"bytes"
	"io"
	"testing"
)

func dump(t *testing.T, r *Ring) {
	t.Logf("%v %v %v %v\n", r.mask, r.tail, r.head, r.data)
}

// Returns the content of the buffer without modifying it
func ringContent(r *Ring) []byte {
	buf := make([]byte, r.Len())
	s := r.end() - r.tail
	if r.head >= r.tail {
		s = r.head - r.tail
	}
	copy(buf, r.data[r.tail:r.tail+s])
	if r.head < r.tail {
		if int(r.head+s) != r.Len() {
			panic("ringContent is wrong (1)")
		}
		copy(buf[s:], r.data[0:r.head])
	} else {
		if int(s) != r.Len() {
			panic("ringContent is wrong (2)")
		}
	}
	return buf
}

func verifyNonMutate(t *testing.T, msg string, truth []byte, ring *Ring) {
	actual := ringContent(ring)
	if !bytes.Equal(actual, truth) {
		t.Errorf("verifyNonMutate failed (%v)", msg)
		t.Logf("truth(%v) : %v", len(truth), truth)
		t.Logf("actual(%v): %v", len(actual), actual)
	}
}

func verify(t *testing.T, useDirectRead bool, truth []byte, ring *Ring) {
	if useDirectRead {
		s1 := ring.DirectRead(len(truth))
		if !bytes.Equal(truth[:len(s1)], s1) {
			t.Error("DirectRead invalid (1)")
		}
		s2 := ring.DirectRead(len(truth))
		if !bytes.Equal(truth[len(s1):len(s1)+len(s2)], s2) {
			t.Error("DirectRead invalid (2)")
		}
	} else {
		buf := make([]byte, len(truth))
		n, err := ring.Read(buf)
		if n != len(truth) {
			t.Errorf("ring.Read returned invalid n. Expected %v, got %v", len(truth), n)
		}
		if err != nil {
			t.Errorf("ring.Read returned unexpected error (expected nil, got %v)", err)
		}
		if !bytes.Equal(truth, buf) {
			t.Error("ring.Read returned invalid data")
		}
	}

	buf := make([]byte, 5)
	n, err := ring.Read(buf)
	if n != 0 || err != io.EOF {
		t.Error("ring.Read is not returned correct values when buffer is empty")
	}
}

func makeTruth() []byte {
	truth := make([]byte, 0, 3000)
	for i := 0; i < cap(truth); i++ {
		truth = append(truth, byte((i+1)%251))
	}
	return truth
}

func TestRingBuffer(t *testing.T) {
	truth := makeTruth()

	for useDirectRead := 0; useDirectRead < 2; useDirectRead++ {
		for len := 1; len < DefaultSize*3; len = len*2 + 1 {
			r := &Ring{}
			copy(r.DirectWrite(len), truth[0:len])
			verify(t, useDirectRead == 1, truth[0:len], r)
		}
	}
}

func TestDirectWrite(t *testing.T) {
	// Stress the branch inside DirectWrite
	r := &Ring{}
	r.DirectWrite(DefaultSize - 5)
	r.DirectRead(DefaultSize - 5)
	// we now have a ring that has storage for DefaultSize bytes, but is 5 bytes away from wrapping around
	b1 := r.DirectWrite(9)
	if len(b1) != 5 {
		t.Errorf("DirectWrite wraparound failed. Expected len %v, got %v", 5, len(b1))
	}
	b2 := r.DirectWrite(4)
	if len(b2) != 4 {
		t.Errorf("DirectWrite wraparound failed. Expected len %v, got %v. %v,%v,%v", 4, len(b2), r.mask, r.tail, r.head)
	}
}

func TestSanity(t *testing.T) {
	if DefaultSize > 128 {
		t.Fatal("Many tests are written with a fixed write size of 100, which assumes that the buffer overflows at 128. You must raise those constants to be closer to DefaultSize")
	}
}

func TestDirectRead(t *testing.T) {
	// stress the "numBytes > int(r.end()-r.tail)" branch in DirectRead
	truth := makeTruth()
	r := &Ring{}

	r.Write(truth[:100])
	verifyNonMutate(t, "0:100", truth[:100], r)

	r.DirectRead(90)
	verifyNonMutate(t, "90:100", truth[90:100], r)

	r.Write(truth[100:150])
	verifyNonMutate(t, "90:150", truth[90:150], r)

	// tail is now at 90, and head is at 90 + 50 - 128 = 12.
	// There are 60 remaining bytes in the buffer.
	b1 := r.DirectRead(60)
	b2 := r.DirectRead(60)
	if len(b1) != 128-90 {
		t.Errorf("DirectRead over edge failed (a) (expected len %v, got %v)", 128-90, len(b1))
	}
	if len(b2) != 60-len(b1) {
		t.Errorf("DirectRead over edge failed (b) (expected len %v, got %v)", 60-len(b1), len(b2))
	}
}

func TestGrowOverEdge(t *testing.T) {
	// Stress the non-trivial growth case
	truth := makeTruth()
	r := &Ring{}
	r.Write(truth[0:100])
	r.DirectRead(100)
	r.Write(truth[100:150])
	verifyNonMutate(t, "Added 100:150", truth[100:150], r)
	// tail is now at 100. Head is at 22 (150-128)
	c1 := ringContent(r)
	cap1 := len(r.data)
	r.Write(truth[:300])
	if len(r.data) == cap1 {
		t.Error("Expected buffer to grow")
	}
	c2 := ringContent(r)
	if r.Len() != len(c1)+300 {
		t.Error("non-trivial growth caused buffer corruption (wrong size)")
	}
	if !bytes.Equal(c1, c2[0:len(c1)]) {
		t.Error("non-trivial growth caused buffer corruption (wrong content)")
	}
}
