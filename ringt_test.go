package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type obj struct {
	id int
}

func TestRingT(t *testing.T) {
	var val []*obj
	var ring RingT[obj]

	nextID := 100

	init := func(maxSize int) {
		val = []*obj{}
		ring = NewRingT[obj](maxSize)
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
			require.Nil(t, actualT)
		}
	}

	add := func() {
		t := &obj{
			id: nextID,
		}
		nextID++
		for len(val) == ring.MaxSize() {
			val = val[1:]
		}
		val = append(val, t)
		ring.Add(t)
	}

	chomp := func() {
		expectEmpty := len(val) == 0
		actual := ring.Next()
		if expectEmpty {
			require.Nil(t, actual)
		} else {
			require.Equal(t, val[0], actual)
			val = val[1:]
		}
	}

	t.Logf("empty")
	init(5)
	validate()

	t.Logf("add 1 at a time")
	for maxSize := 1; maxSize < 7; maxSize++ {
		init(maxSize)
		for i := 0; i < 20; i++ {
			add()
			validate()
		}
	}

	init(5)
	add()
	add()
	add()
	validate()
	for i := 0; i < 5; i++ {
		chomp()
	}
	validate()
}
