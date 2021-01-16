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
			So(len(w.ComponentFlags), ShouldEqual, 0)
			w.RemoveEntity(0)
			So(len(w.ComponentFlags), ShouldEqual, 0)
		})

		Convey("An entity can be removed", func() {
			e := w.NewEntity()

			w.RemoveEntity(e)

			So(len(w.ComponentFlags), ShouldEqual, 0)
		})

		Convey("An entity can be removed more than once", func() {
			e := w.NewEntity()

			w.RemoveEntity(e)
			w.RemoveEntity(e)
			w.RemoveEntity(e)

			So(len(w.ComponentFlags), ShouldEqual, 0)
		})
	})
}