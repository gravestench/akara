package akara

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestComponentFactory_Add_Get_Remove(t *testing.T) {
	rand.Seed(int64(0xdeadbeef))

	Convey("Within an ECS world", t, func() {
		cfg := NewWorldConfig().With(nil)
		w := NewWorld(cfg)
		var e EID

		Convey("Where there exists a registered component type with Component ID 0", func() {
			cid := w.RegisterComponent(&testComponent{})
			cf := w.GetComponentFactory(cid)

			Convey("For a given entity", func() {
				e = w.NewEntity()

				Convey("The entity does not implicitly have a component of type 0", func() {
					_, found := cf.Get(e)
					So(found, ShouldEqual, false)
				})

				c := cf.Add(e)
				Convey("The entity can be given a component of type 0", func() {
					So(c, ShouldNotBeNil)
				})

				Convey("The component factory does not clobber existing component instances", func() {
					c2 := cf.Add(e)
					So(c, ShouldEqual, c2)
				})

				Convey("The component can be removed", func() {
					cf.Remove(e)

					_, found := cf.Get(e)

					So(found, ShouldEqual, false)
				})
			})
		})
	})
}

func benchComMapAdd(_ int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	cfg := NewWorldConfig().With(nil)
	w := NewWorld(cfg)
	cid := w.RegisterComponent(&testComponent{})
	cf := w.GetComponentFactory(cid)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cf.Add(w.NewEntity())
	}
}

func benchComMapGet(i int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	cfg := NewWorldConfig().With(nil)
	w := NewWorld(cfg)
	cid := w.RegisterComponent(&testComponent{})
	cf := w.GetComponentFactory(cid)

	for n := 0; n < i; n++ {
		cf.Add(w.NewEntity())
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cf.Get(EID(n%i))
	}
}

func benchComMapRemove(i int, b *testing.B) {
	rand.Seed(int64(0xdeadbeef))

	cfg := NewWorldConfig().With(nil)
	w := NewWorld(cfg)
	cf := w.GetComponentFactory(w.RegisterComponent(&testComponent{}))

	randNums := randIndices(i)
	eids := make([]EID, len(randNums))

	for idx := range randNums {
		eids[idx] = w.NewEntity()
		cf.Add(eids[idx])
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, eid := range eids {
			cf.Remove(eid)
		}
	}
}

func BenchmarkComponentFactory_AddAll(b *testing.B) {
	for i := 2; i < 7; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("%d entries", v), func(b *testing.B) {
			benchComMapAdd(v, b)
		})
	}
}

func BenchmarkComponentFactory_GetAll(b *testing.B) {
	for i := 2; i < 7; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("%d entries", v), func(b *testing.B) {
			benchComMapGet(v, b)
		})
	}
}

func BenchmarkComponentFactory_RemoveAll(b *testing.B) {
	for i := 2; i < 7; i++ {
		v := int(math.Pow10(i))
		b.Run(fmt.Sprintf("%d entries", v), func(b *testing.B) {
			benchComMapRemove(v, b)
		})
	}
}

type testComponent struct {}

func (c *testComponent) New() Component {
	return &testComponent{}
}
