package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type pod struct {
	id int
}

func TestRingP(t *testing.T) {
	var val []pod
	var ring RingP[pod]
	var zero pod

	nextID := 100

	init := func(maxSize int) {
		val = []pod{}
		ring = NewRingP[pod](maxSize)
	}

	validate := func() {
		require.Equal(t, len(val), ring.Len())
		for i := 0; i < len(val); i++ {
			actualT := ring.Peek(i)
			expectT := val[i]
			require.Equal(t, expectT, actualT)
		}
		// verify that Peek(<invalid index>) returns nil
		invalidIndices := []int{-1, len(val), len(val) + 1}
		for _, invalidI := range invalidIndices {
			actualT := ring.Peek(invalidI)
			require.Equal(t, actualT, zero)
		}
	}

	add := func() {
		t := pod{
			id: nextID,
		}
		nextID++
		for len(val) == ring.Capacity() {
			val = val[1:]
		}
		val = append(val, t)
		ring.Add(t)
	}

	chomp := func() {
		expectEmpty := len(val) == 0
		actual := ring.Next()
		if expectEmpty {
			require.Equal(t, actual, zero)
		} else {
			require.Equal(t, val[0], actual)
			val = val[1:]
		}
	}

	t.Logf("empty")
	init(4)
	validate()

	t.Logf("add 1 at a time")
	for size := 2; size < 32; size *= 2 {
		init(size)
		for i := 0; i < 20; i++ {
			add()
			validate()
		}
	}

	init(4)
	add()
	add()
	add()
	validate()
	for i := 0; i < 5; i++ {
		chomp()
	}
	validate()
}
