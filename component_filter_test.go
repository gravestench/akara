package akara

import "testing"

func Test_ComponentFilter_WithBuilder(t *testing.T) {
	bs1 := NewBitSet(1, 2, 3)
	bs2 := NewBitSet(1, 2, 3, 4)

	cf1 := NewComponentFilter(bs1, nil, nil)

	if !cf1.Allow(bs2) {
		t.Error("component filter should allow bitset which is superset of it's Required bitset")
	}
}
