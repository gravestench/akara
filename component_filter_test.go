package akara

import (
	"fmt"
	"github.com/gravestench/bitset"
	"math"
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func randIndices(max int) []int {
	count := max
	ids := make([]int, count)

	for idx := range ids {
		if rand.Intn(2) == 0 {
			// coin toss
			continue
		}

		ids[idx] = rand.Intn(max)
	}

	return ids
}

func TestComponentFilter_Allow(t *testing.T) {
	Convey("With BitSets A and B, Where true bits in A are a subset of the true bits in B", t, func() {
		A := bitset.NewBitSet(1, 2, 3)
		B := bitset.NewBitSet(1, 2, 3, 4)

		Convey("Where the filter imposes no requirements or restrictions", func() {
			cf := NewComponentFilter(nil, nil, nil)

			Convey("The filter will allow any bitset", func() {
				So(cf.Allow(A), ShouldEqual, true)
				So(cf.Allow(B), ShouldEqual, true)
			})
		})

		Convey("With a ComponentFilter that requires all bits from BitSet A", func() {
			cf1 := NewComponentFilter(A, nil, nil)

			Convey("the component filter should allow bitset A", func() {
				So(cf1.Allow(A), ShouldEqual, true)
			})

			Convey("the component filter should allow bitset B", func() {
				So(cf1.Allow(B), ShouldEqual, true)
			})
		})

		Convey("With a ComponentFilter that requires all bits from BitSet B", func() {
			cf1 := NewComponentFilter(B, nil, nil)

			Convey("the component filter should NOT allow bitset A", func() {
				So(cf1.Allow(A), ShouldEqual, false)
			})

			Convey("the component filter should allow bitset B", func() {
				So(cf1.Allow(B), ShouldEqual, true)
			})
		})
	})
}

func TestComponentFilter_Equals(t *testing.T) {
	bs1 := bitset.NewBitSet(1, 2, 3)
	bs2 := bitset.NewBitSet(1, 2, 3, 4)

	cf1 := NewComponentFilter(bs1, nil, nil)

	if !cf1.Allow(bs2) {
		t.Error("component filter should allow bitset which is superset of it's Required bitset")
	}
}

func benchComponentFilterAllow(i int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	// this setup can be expensive, we do this first...
	bs1 := bitset.NewBitSet(randIndices(i)...)
	bs2 := bitset.NewBitSet(randIndices(i)...)
	bs3 := bitset.NewBitSet(randIndices(i)...)
	cf1 := NewComponentFilter(bs1, bs2, bs3)

	// ... then we need to reset the timer...
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cf1.Allow(bitset.NewBitSet(randIndices(i)...))
	}
}

func BenchmarkComponentFilter_Allow(b *testing.B) {
	for i := 2; i < 7; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchComponentFilterAllow(v, b)
		})
	}
}

func benchComponentFilterEquals(i int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	bs1 := bitset.NewBitSet(randIndices(i)...)
	bs2 := bitset.NewBitSet(randIndices(i)...)
	bs3 := bitset.NewBitSet(randIndices(i)...)
	bs4 := bitset.NewBitSet(randIndices(i)...)
	bs5 := bitset.NewBitSet(randIndices(i)...)
	bs6 := bitset.NewBitSet(randIndices(i)...)

	cf1 := NewComponentFilter(bs1, bs2, bs3)
	cf2 := NewComponentFilter(bs4, bs5, bs6)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cf1.Equals(cf2)
	}
}

func BenchmarkComponentFilter_Equals(b *testing.B) {
	for i := 2; i < 7; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("%d bits", v), func(b *testing.B) {
			benchComponentFilterEquals(v, b)
		})
	}
}
