package akara

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestComponentFactory_Add(t *testing.T) {

}

func TestComponentFactory_Get(t *testing.T) {

}

func TestComponentFactory_Remove(t *testing.T) {

}

func benchComMapAdd(i int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	cf := newComponentFactory(0)
	cf.provider = func() Component {return &testComponent{}}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for idx := range randIndices(i) {
			cf.Add(EID(idx))
		}
	}
}

func benchComMapGet(i int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	cf := newComponentFactory(0)
	cf.provider = func() Component {return &testComponent{}}

	for idx := range randIndices(i) {
		cf.Add(EID(idx))
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for idx := range randIndices(i) {
			cf.Get(EID(idx))
		}
	}
}

func benchComMapRemove(i int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	cf := newComponentFactory(0)
	cf.provider = func() Component {return &testComponent{}}

	for idx := range randIndices(i) {
		cf.Add(EID(idx))
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for idx := range randIndices(i) {
			cf.Remove(EID(idx))
		}
	}
}

func BenchmarkComponentFactory_Add(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("ComponentFactory.Add %d entries", v), func(b *testing.B) {
			benchComMapAdd(v, b)
		})
	}
}

func BenchmarkComponentFactory_Get(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("ComponentFactory.Get %d entries", v), func(b *testing.B) {
			benchComMapGet(v, b)
		})
	}
}

func BenchmarkComponentFactory_Remove(b *testing.B) {
	for i := 2; i < 10; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("ComponentFactory.Remove %d entries", v), func(b *testing.B) {
			benchComMapRemove(v, b)
		})
	}
}

type testComponent struct {}

func (c *testComponent) New() Component {
	return &testComponent{}
}
