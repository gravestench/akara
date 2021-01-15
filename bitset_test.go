package akara

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBitSet_Get(t *testing.T) {
	Convey("With a new BitSet", t, func() {
		bs1 := NewBitSet()

		Convey("The bit at any index should be false", func() {
			for i := 0; i < 100; i++ {
				So(bs1.Get(i), ShouldEqual, false)
			}
		})
	})

	Convey("With a new BitSet, where index 1 is set to true", t, func() {
		bs1 := NewBitSet(1)

		Convey("Only bit index 1 should be true", func() {
			for i := 0; i < 100; i++ {
				So(bs1.Get(i), ShouldEqual, i==1)
			}
		})
	})
}

func TestBitSet_Or(t *testing.T) {
	Convey("With two BitSets, A and B", t, func() {
		Convey("Where A is all 0's, and B has a random bit index set to true", func() {
			A := NewBitSet()
			B := NewBitSet(rand.Intn(1000))

			Convey("A|B should be true", func() {
				So(!(A.Clone().Or(B).Empty()), ShouldEqual, true)
			})

			Convey("B|A should be true", func() {
				So(!(B.Clone().Or(A).Empty()), ShouldEqual, true)
			})
		})

		Convey("Where both A and B are all 0's", func() {
			B := NewBitSet()
			A := NewBitSet()

			Convey("A|B should be false", func() {
				So(!(A.Clone().Or(B).Empty()), ShouldEqual, false)
			})

			Convey("B|A should be false", func() {
				So(!(B.Clone().Or(A).Empty()), ShouldEqual, false)
			})
		})

		Convey("Where both A and B have random indices set to true", func() {
			B := NewBitSet(rand.Intn(1000))
			A := NewBitSet(rand.Intn(1000))

			Convey("A|B should be true", func() {
				So(!(A.Clone().Or(B).Empty()), ShouldEqual, true)
			})

			Convey("B|A should be true", func() {
				So(!(B.Clone().Or(A).Empty()), ShouldEqual, true)
			})
		})
	})
}

func TestBitSet_And(t *testing.T) {
	Convey("With two BitSets, A and B", t, func() {
		Convey("Where A is all 0's, and B has a random bit index set to true", func() {
			A := NewBitSet()
			B := NewBitSet(rand.Intn(1000))

			Convey("A&B should be false", func() {
				So(A.Clone().And(B).Empty(), ShouldEqual, true)
			})

			Convey("B&A should be false", func() {
				So(!(B.Clone().And(A).Empty()), ShouldEqual, false)
			})
		})

		Convey("Where both A and B are all 0's", func() {
			B := NewBitSet()
			A := NewBitSet()

			Convey("A&B should be false", func() {
				So(!(A.Clone().And(B).Empty()), ShouldEqual, false)
			})

			Convey("B&A should be false", func() {
				So(!(B.Clone().And(A).Empty()), ShouldEqual, false)
			})
		})

		Convey("Where both A and B have the same random index set to true", func() {
			idx := rand.Intn(1000)
			B := NewBitSet(idx)
			A := NewBitSet(idx)

			Convey("A&B should be true", func() {
				So(!(A.Clone().And(B).Empty()), ShouldEqual, true)
			})

			Convey("B&A should be true", func() {
				So(!(B.Clone().And(A).Empty()), ShouldEqual, true)
			})
		})
	})
}

func TestBitSet_ContainsAll(t *testing.T) {
	Convey("With two BitSets, A and B", t, func() {
		Convey("Where A and B have no true bits", func() {
			A := NewBitSet()
			B := NewBitSet()

			Convey("B ∈ A (B is contained by A) should be true", func() {
				So(A.ContainsAll(B), ShouldEqual, true)
			})

			Convey("A ∈ B (A is contained by B) should be true", func() {
				v := B.ContainsAll(A)
				So(v, ShouldEqual, true)
			})
		})

		Convey("Where A has a single true bit, B has none", func() {
			A := NewBitSet(0)
			B := NewBitSet()

			Convey("B ∈ A (B is contained by A) should be true", func() {
				So(A.ContainsAll(B), ShouldEqual, true)
			})

			Convey("A ∈ B (A is contained by B) should be false", func() {
				v := B.ContainsAll(A)
				So(v, ShouldEqual, false)
			})
		})

		Convey("Where A and B both share a single true bit (same bit index)", func() {
			A := NewBitSet(1)
			B := NewBitSet(1)

			Convey("B ∈ A (B is contained by A) should be true", func() {
				So(A.ContainsAll(B), ShouldEqual, true)
			})

			Convey("A ∈ B (A is contained by B) should be false", func() {
				v := B.ContainsAll(A)
				So(v, ShouldEqual, true)
			})
		})

		Convey("Where A and B do not have a true bit in common", func() {
			A := NewBitSet(0)
			B := NewBitSet(1)

			Convey("B ∈ A (B is contained by A) should be true", func() {
				So(A.ContainsAll(B), ShouldEqual, false)
			})

			Convey("A ∈ B (A is contained by B) should be false", func() {
				v := B.ContainsAll(A)
				So(v, ShouldEqual, false)
			})
		})
	})
}

func TestBitSet_Intersects(t *testing.T) {
	Convey("With two BitSets, A and B", t, func() {
		Convey("Where A and B have no true bits", func() {
			A := NewBitSet()
			B := NewBitSet()

			Convey("A ∩ B (A intersected by B) should be false", func() {
				v := B.Intersects(A)

				So(v, ShouldEqual, false)
			})

			Convey("B ∩ A (B intersected by A) should be false", func() {
				So(A.Intersects(B), ShouldEqual, false)
			})
		})

		Convey("Where A and B have no true bits in common", func() {
			A := NewBitSet(0)
			B := NewBitSet(1)

			Convey("A ∩ B (A intersected by B) should be false", func() {
				v := B.Intersects(A)

				So(v, ShouldEqual, false)
			})

			Convey("B ∩ A (B intersected by A) should be false", func() {
				So(A.Intersects(B), ShouldEqual, false)
			})
		})

		Convey("Where A and B have only one true bit in common", func() {
			A := NewBitSet(0, 1)
			B := NewBitSet(1, 2)

			Convey("A ∩ B (A intersected by B) should be true", func() {
				v := B.Intersects(A)

				So(v, ShouldEqual, true)
			})

			Convey("B ∩ A (B intersected by A) should be true", func() {
				So(A.Intersects(B), ShouldEqual, true)
			})
		})

		Convey("Where A has true bits and B has no true bits", func() {
			A := NewBitSet(0, 1)
			B := NewBitSet()

			Convey("A ∩ B (A is contained by B) should be false", func() {
				v := B.Intersects(A)

				So(v, ShouldEqual, false)
			})

			Convey("B ∩ A (B is contained by A) should be false", func() {
				So(A.Intersects(B), ShouldEqual, false)
			})
		})
	})
}

func TestBitSet_Equals(t *testing.T) {
	Convey("With two BitSets, A and B", t, func() {
		Convey("Where A and B have no bits set true", func() {
			A := NewBitSet()
			B := NewBitSet()

			Convey("A == B should be true", func() {
				So(A.Equals(B), ShouldEqual, true)
			})

			Convey("B == A should be true", func() {
				v := B.Equals(A)

				So(v, ShouldEqual, true)
			})
		})

		Convey("Where A and B share true bits", func() {
			A := NewBitSet(0)
			B := NewBitSet(0)

			Convey("A == B should be true", func() {
				So(A.Equals(B), ShouldEqual, true)
			})

			Convey("B == A should be true", func() {
				v := B.Equals(A)

				So(v, ShouldEqual, true)
			})
		})

		Convey("Where A and B are not identical", func() {
			A := NewBitSet(0)
			B := NewBitSet(1)

			Convey("A == B should be false", func() {
				So(A.Equals(B), ShouldEqual, false)
			})

			Convey("B == A should be false", func() {
				v := B.Equals(A)

				So(v, ShouldEqual, false)
			})
		})
	})
}

func TestBitSet_ToIntArray(t *testing.T) {
	Convey("Given a bitset", t, func() {
		Convey("After clearing the bitset", func() {
			Convey("ToIntArray should yield no indices", func() {
				bs := NewBitSet()

				So(bs.ToIntArray(), ShouldBeEmpty)
			})
		})

		Convey("After setting a single bit in the bitset to true", func() {
			Convey("ToIntArray should yield one index", func() {
				bs := NewBitSet()

				bs.Set(0, true)
				So(len(bs.ToIntArray()), ShouldEqual, 1)
			})
		})

		Convey("After setting a single bit in the bitset to true, twice", func() {
			Convey("ToIntArray should yield one index", func() {
				bs := NewBitSet()

				bs.Set(0, true)
				So(len(bs.ToIntArray()), ShouldEqual, 1)
			})
		})

		Convey("After setting N bits to true", func() {
			Convey("ToIntArray should yield N indices", func() {
				bs := NewBitSet()

				n := rand.Intn(100)
				for idx := 0; idx < n; idx++ {
					bs.Set(idx, true)
				}

				So(len(bs.ToIntArray()), ShouldEqual, n)
			})
		})
		Convey("", func() {})
	})
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
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchBitsetGet(v, b)
		})
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
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchBitsetOr(v, b)
		})
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
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchBitsetAnd(v, b)
		})
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
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchBitsetContainsAll(v, b)
		})
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
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchBitsetIntersects(v, b)
		})
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
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchBitsetEquals(v, b)
		})
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
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchBitsetToIntArray(v, b)
		})
	}
}
