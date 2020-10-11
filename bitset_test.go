package akara

import "testing"

func TestBitSet_Get(t *testing.T) {
	bs := NewBitSet()

	if bs.Get(0) {
		t.Error("value should be false")
	}

	if bs.Get(1 << 32) {
		t.Error("value should be false")
	}

	bs.Set(0, true)

	if !bs.Get(0) {
		t.Error("value should be true")
	}

	bs.Set(0, false)

	if bs.Get(0) {
		t.Error("value should be false")
	}

	bs.Set(64, true)

	if !bs.Get(64) {
		t.Error("value should be true")
	}
}

func TestBitSet_Or(t *testing.T) {
	bs1 := NewBitSet()
	bs1.Set(8, false)

	bs2 := bs1.Clone()

	bs1.Set(1, true)
	bs1.Set(8, true)
	bs2.Or(bs1)

	if !bs2.Get(1) || !bs2.Get(8) {
		t.Error("Bitset OR operation failed")
	}
}

func TestBitSet_And(t *testing.T) {
	bs1 := NewBitSet()
	bs1.Set(8, false)

	bs2 := bs1.Clone()
	bs2.Invert()

	bs1.Set(1, true)
	bs1.Set(8, true)

	bs2.And(bs1)

	if bs2.Get(0) || !bs2.Get(1) || !bs2.Get(8) {
		t.Error("Bitset AND operation failed")
	}
}

func TestBitSet_ContainsAll(t *testing.T) {
	bs1 := NewBitSet()
	bs1.Set(0, false)

	bs2 := bs1.Clone()

	bs1.Set(0, true)
	bs1.Set(1, true)
	bs1.Set(2, true)
	bs1.Set(65, true)

	bs2.Set(0, true)
	bs2.Set(2, true)

	if !bs1.ContainsAll(bs2) {
		t.Error("super-set contains the sub-set")
	}

	if bs2.ContainsAll(bs1) {
		t.Error("sub-set does not contain super-set")
	}
}

func TestBitSet_Intersects(t *testing.T) {
	bs1 := NewBitSet()
	bs1.Set(0, false)

	bs2 := bs1.Clone()
	bs3 := bs1.Clone()

	bs1.Set(0, true)
	bs1.Set(1, true)
	bs1.Set(2, true)

	bs2.Set(0, true)
	bs2.Set(2, true)

	bs3.Set(65, true)

	if !bs1.Intersects(bs2) || !bs2.Intersects(bs1) {
		t.Error("super-set contains the sub-set")
	}

	if bs3.Intersects(bs1) {
		t.Error("sub-set does not contain super-set")
	}
}

func TestBitSet_Equals(t *testing.T) {
	bs1 := NewBitSet(1, 2, 4, 8)
	bs2 := bs1.Clone()

	if !bs1.Equals(bs2) {
		t.Error("cloned bitset should be equal")
	}

	bs2.Set(0, true)

	if bs1.Equals(bs2) {
		t.Error("unequal bitsets should not be equal")
	}
}

func TestBitSet_ToIntArray(t *testing.T) {
	bs1 := NewBitSet(1, 2, 4, 8, 64)

	a := bs1.ToIntArray()

	if len(a) != 5 {
		t.Error("unexpected indices in result int array")
	}
}
