package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type thing struct {
	id     int
	weight int
}

func TestRingT(t *testing.T) {
	val := []*thing{}
	valW := 0
	ring := NewRingT[thing](10)

	nextID := 0

	clear := func() {
		val = []*thing{}
		valW = 0
		ring = NewRingT[thing](10)
	}

	validate := func() {
		require.Equal(t, len(val), ring.Len())
		require.Equal(t, valW, ring.Weight())
		for i := uint(0); i < uint(len(val)); i++ {
			actualT, actualW := ring.peek(i)
			expectT := val[i]
			expectW := val[i].weight
			require.Equal(t, expectT, actualT)
			require.Equal(t, expectW, actualW)
		}
	}

	add := func(weight int) {
		t := &thing{
			id:     nextID,
			weight: weight,
		}
		for valW+weight > ring.MaxWeight && len(val) != 0 {
			val = val[1:]
			valW -= val[0].weight
		}
		val = append(val, t)
		valW += weight
		ring.Add(weight, t)
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
	for i := 0; i < 30; i++ {
		t.Logf("add %v", i)
		add(i)
		validate()
	}

	clear()
	for i := 0; i < 9; i++ {
		add(6)
		validate()
	}
}
