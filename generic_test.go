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
		_, actual, actualW := ring.Next()
		require.Equal(t, val[0], actual)
		require.Equal(t, val[0].weight, actualW)
		valW -= val[0].weight
		val = val[1:]
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
	for i := 0; i < 3; i++ {
		chomp()
	}
	validate()

}
