package akara

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorld_NewEntity(t *testing.T) {
	Convey("For a given ECS world", t, func() {
		w := NewWorld(NewWorldConfig())

		first := w.NewEntity()

		Convey("The first entity has ID 1", func() {
			So(first, ShouldEqual, 1)
		})

		Convey("Entity ID's always increase in size", func() {
			start := time.Now()
			for time.Since(start) <= time.Millisecond {
				last := w.NewEntity()
				So(last, ShouldBeGreaterThan, first)
				first = last
			}
		})
	})
}

func TestWorld_RemoveEntity(t *testing.T) {
	Convey("For a given ECS world", t, func() {
		w := NewWorld(NewWorldConfig())

		Convey("An entity can always be removed, even if it does not exist", func() {
			_, found := w.ComponentFlags.Load(EntityID(1))
			So(found, ShouldBeFalse)

			w.RemoveEntity(0)

			_, found = w.ComponentFlags.Load(EntityID(1))
			So(found, ShouldBeFalse)
		})

		Convey("An entity can be removed", func() {
			e := w.NewEntity()

			_, found := w.ComponentFlags.Load(EntityID(1))
			So(found, ShouldBeTrue)

			w.RemoveEntity(e)
			w.Update(0)

			_, found = w.ComponentFlags.Load(EntityID(1))
			So(found, ShouldBeFalse)
		})

		Convey("An entity can be removed more than once", func() {
			e := w.NewEntity()

			_, found := w.ComponentFlags.Load(EntityID(1))
			So(found, ShouldBeTrue)

			w.RemoveEntity(e)
			w.RemoveEntity(e)
			w.RemoveEntity(e)
			w.Update(0)

			_, found = w.ComponentFlags.Load(EntityID(1))
			So(found, ShouldBeFalse)
		})

		Convey("An entity which is removed is also removed from all subscriptions", func() {
			cid := w.RegisterComponent(&testComponent{})
			testComponentFactory := &testComponentFactory{ComponentFactory: w.GetComponentFactory(cid)}

			sub := w.AddSubscription(w.NewComponentFilter().Require(&testComponent{}).Build())

			e := w.NewEntity()
			testComponentFactory.Add(e)

			So(len(sub.GetEntities()), ShouldEqual, 1)

			_, found := w.ComponentFlags.Load(e)
			So(found, ShouldBeTrue)

			w.RemoveEntity(e)
			w.Update(0)

			So(len(sub.GetEntities()), ShouldEqual, 0)

			_, found = w.ComponentFlags.Load(e)
			So(found, ShouldBeFalse)
		})
	})
}

type testComponentFactory struct {
	*ComponentFactory
}

func (m *testComponentFactory) Add(id EID) *testComponent {
	return m.ComponentFactory.Add(id).(*testComponent)
}

func (m *testComponentFactory) Get(id EID) (*testComponent, bool) {
	component, found := m.ComponentFactory.Get(id)
	if !found {
		return nil, found
	}

	return component.(*testComponent), found
}

