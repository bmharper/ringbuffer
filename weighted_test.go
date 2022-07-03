package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type thing struct {
	id     int
	weight int
}

func TestWeightedRingT(t *testing.T) {
	val := []*thing{}
	valW := 0
	ring := NewWeightedRingT[thing](10)

	nextID := 0

	clear := func() {
		val = []*thing{}
		valW = 0
		ring = NewWeightedRingT[thing](10)
	}

	validate := func() {
		require.Equal(t, len(val), ring.Len())
		require.Equal(t, valW, ring.Weight())
		for i := 0; i < len(val); i++ {
			actualExist, actualT, actualW := ring.Peek(i)
			expectT := val[i]
			expectW := val[i].weight
			require.Equal(t, true, actualExist)
			require.Equal(t, expectT, actualT)
			require.Equal(t, expectW, actualW)
		}
		// verify that Peek(<invalid index>) returns 'no item'
		invalidIndices := []int{-1, len(val), len(val) + 1}
		for _, invalidI := range invalidIndices {
			actualExist, actualT, actualW := ring.Peek(invalidI)
			require.Equal(t, false, actualExist)
			require.Nil(t, actualT)
			require.Equal(t, 0, actualW)
		}
	}

	add := func(weight int) {
		t := &thing{
			id:     nextID,
			weight: weight,
		}
		nextID++
		for valW+weight > ring.MaxWeight && len(val) != 0 {
			valW -= val[0].weight
			val = val[1:]
		}
		val = append(val, t)
		valW += weight
		ring.Add(weight, t)
	}

	chomp := func() {
		expectEmpty := len(val) == 0
		ok, actual, actualW := ring.Next()
		if expectEmpty {
			require.Equal(t, false, ok)
			require.Nil(t, actual)
			require.Equal(t, actualW, 0)
		} else {
			require.Equal(t, val[0], actual)
			require.Equal(t, val[0].weight, actualW)
			valW -= val[0].weight
			val = val[1:]
		}
	}

	t.Logf("empty")
	validate()

	t.Logf("add 1 at a time")
	for i := 0; i < 20; i++ {
		add(1)
		validate()
	}

	clear()
	t.Logf("add i at a time")
	for i := 0; i < 50; i++ {
		//t.Logf("add %v", i)
		add(i % (ring.MaxWeight + 1))
		validate()
	}

	clear()
	for i := 0; i < 3; i++ {
		add(9)
		validate()
	}

	clear()
	add(2)
	add(3)
	add(4)
	validate()
	for i := 0; i < 5; i++ {
		chomp()
	}
	validate()
}
