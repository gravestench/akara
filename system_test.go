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

func TestSystem(t *testing.T) {
	var w *World
	var sys *MyTestSystem

	Convey("Given an ECS World with a system", t, func() {
		w = NewWorld(NewWorldConfig())

		So(len(w.Systems), ShouldEqual, 0)

		sys = &MyTestSystem{}
		w.AddSystem(sys)

		So(len(w.Systems), ShouldEqual, 1)

		Convey("The System does not tick unless the World ticks", func() {
			So(sys.Ticks, ShouldEqual, 0)
		})

		Convey("An unaltered System has the default tick rate", func() {
			So(sys.TickRate(), ShouldEqual, DefaultTickRate)
		})

		Convey("The System will tick at its target tick rate", func() {
			tickDuration := time.Duration(float64(time.Second) * (1 / sys.TickRate()))
			w.Update(tickDuration)
			w.Update(tickDuration)
			w.Update(tickDuration)

			So(sys.Ticks, ShouldEqual, 3)
		})

		Convey("The System will tick slower than its target tick rate", func() {
			tickDuration := time.Duration(float64(time.Second) * (1 / sys.TickRate()))
			tickDuration += 1 // arbitrarily longer than the target tick rate
			w.Update(tickDuration)
			w.Update(tickDuration)
			w.Update(tickDuration)

			So(sys.Ticks, ShouldEqual, 3)
		})

		Convey("The System will not tick faster than its target tick rate", func() {
			tickDuration := time.Duration(float64(time.Second) * (1 / sys.TickRate()))
			tickDuration -= 1 // arbitrarily shorter than the target tick rate
			w.Update(tickDuration)
			w.Update(tickDuration)
			w.Update(tickDuration)

			// if target tick rate is once per 1000ns max, and we try to tick at 999ns, 1998ns, and 2997ns,
			// 999ns == no tick (last tick 999ns ago, 999ns < 1000ns so we do not tick)
			// 1998ns == one tick (last tick 1998ns ago, 1998ns >= 1000ns so we tick)
			// 2997ns == no tick (last tick 999ns ago)
			// so we expect one tick here
			So(sys.Ticks, ShouldEqual, 1)
		})

		Convey("We can change the System's tick rate", func() {
			So(sys.TickRate(), ShouldEqual, DefaultTickRate)
			sys.SetTickRate(999)
			So(sys.TickRate(), ShouldEqual, 999)
		})
	})
}
