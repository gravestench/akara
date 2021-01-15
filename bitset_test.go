package akara

import (
	"fmt"
	"math"
	"testing"
)

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

func benchBitsetGet(i int, b *testing.B) {
	bs1 := NewBitSet(i)
	bs1.Set(i, true)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bs1.Get(i)
	}
}

func BenchmarkBitSet_Get(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("Bitset.Get %d bits", v), func(b *testing.B) {
			benchBitsetGet(v, b)
		})
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

func benchBitsetOr(i int, b *testing.B) {
	bs1 := NewBitSet(i)
	bs2 := NewBitSet(i)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bs1.Or(bs2)
	}
}

func BenchmarkBitSet_Or(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("Bitset.Or %d bits", v), func(b *testing.B) {
			benchBitsetOr(v, b)
		})
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

func benchBitsetAnd(i int, b *testing.B) {
	bs1 := NewBitSet(i)
	bs2 := NewBitSet(i)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bs1.And(bs2)
	}
}

func BenchmarkBitSet_And(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("Bitset.And %d bits", v), func(b *testing.B) {
			benchBitsetAnd(v, b)
		})
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

func benchBitsetContainsAll(i int, b *testing.B) {
	bs1 := NewBitSet(i)
	bs2 := NewBitSet(i)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bs1.ContainsAll(bs2)
	}
}

func BenchmarkBitSet_ContainsAll(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("Bitset.ContainsAll %d bits", v), func(b *testing.B) {
			benchBitsetContainsAll(v, b)
		})
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

func benchBitsetIntersects(i int, b *testing.B) {
	bs1 := NewBitSet(i)
	bs2 := NewBitSet(i)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bs1.Intersects(bs2)
	}
}

func BenchmarkBitSet_Intersects(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("Bitset.Intersects %d bits", v), func(b *testing.B) {
			benchBitsetIntersects(v, b)
		})
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

func benchBitsetEquals(i int, b *testing.B) {
	bs1 := NewBitSet(i)
	bs2 := NewBitSet(i)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bs1.Equals(bs2)
	}
}

func BenchmarkBitSet_Equals(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("Bitset.Equals %d bits", v), func(b *testing.B) {
			benchBitsetEquals(v, b)
		})
	}
}

func TestBitSet_ToIntArray(t *testing.T) {
	bs1 := NewBitSet(1, 2, 4, 8, 64)

	a := bs1.ToIntArray()

	if len(a) != 5 {
		t.Error("unexpected indices in result int array")
	}
}

func benchBitsetToIntArray(i int, b *testing.B) {
	bs := NewBitSet(i)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bs.ToIntArray()
	}
}

func BenchmarkBitSet_ToIntArray(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("Bitset.ToIntArray %d bits", v), func(b *testing.B) {
			benchBitsetToIntArray(v, b)
		})
	}
}
