package akara

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type MyTestSystem struct {
	BaseSystem
	Ticks int
}

func (sys *MyTestSystem) Update() {
	sys.Ticks += 1
}

func TestBaseSystem_TickFrequency(t *testing.T) {
	Convey("Given a System", t, func() {
		sys := &MyTestSystem{}

		Convey("I can set the tick frequency", func() {
			sys.SetTickFrequency(100)
			So(sys.TickFrequency(), ShouldEqual, 100)

			Convey("The tick period is calculated correctly", func() {
				So(sys.TickPeriod(), ShouldEqual, 10*time.Millisecond)
			})
		})
	})
}

func TestSystem(t *testing.T) {
	var w *World
	var sys *MyTestSystem

	Convey("Given an ECS World", t, func() {
		w = NewWorld()

		Convey("A World starts with no Systems", func() {
			So(len(w.Systems), ShouldEqual, 0)
		})

		Convey("I can add a System to the World without activating it", func() {
			sys = &MyTestSystem{}
			activate := false
			w.AddSystem(sys, activate)

			So(len(w.Systems), ShouldEqual, 1)

			w.Update()

			So(sys.Active(), ShouldBeFalse)

			Convey("An unaltered System has the default tick rate", func() {
				So(sys.TickFrequency(), ShouldEqual, DefaultTickRate)
			})

			Convey("We can change the System's tick rate", func() {
				So(sys.TickFrequency(), ShouldEqual, DefaultTickRate)
				sys.SetTickFrequency(999)
				So(sys.TickFrequency(), ShouldEqual, 999)
			})

			Convey("An unactivated System does not tick on its own", func() {
				time.Sleep(50 * time.Millisecond) // arbitrary sleep, the system doesn't tick no matter how long we wait
				So(sys.Ticks, ShouldEqual, 0)
			})

			Convey("I can manually tick an unactivated System", func() {
				sys.Tick()
				sys.Tick()
				sys.Tick()
				So(sys.Ticks, ShouldEqual, 3)
			})
		})

		Convey("I can add a System to the World and activate it", func() {
			sys = &MyTestSystem{}
			activate := true
			w.AddSystem(sys, activate)

			So(len(w.Systems), ShouldEqual, 1)

			// the system becomes active on the next world update
			So(sys.Active(), ShouldBeFalse)
			w.Update()
			So(sys.Active(), ShouldBeTrue)

			Convey("An activated System ticks on its own", func() {
				time.Sleep(50 * time.Millisecond) // arbitrary sleep to get some ticks in
				So(sys.Ticks, ShouldBeGreaterThan, 0)
			})

			Convey("I can deactivate a System", func() {
				sys.Deactivate()
				So(sys.Active(), ShouldBeFalse)

				Convey("A deactivated System does not tick automatically", func() {
					oldTickCount := sys.Ticks
					time.Sleep(50 * time.Millisecond)
					So(sys.Ticks, ShouldEqual, oldTickCount)
				})
			})
		})
	})
}
