package akara

import (
	"testing"

	"github.com/gravestench/bitset"

	. "github.com/smartystreets/goconvey/convey"
)

func TestComponentFilterBuilder(t *testing.T) {
	Convey("Given an Arbitrary BitSet", t, func() {
		bs := bitset.NewBitSet()

		Convey("A component filter may be created with any combination of requirements/restrictions", func () {
			So(NewComponentFilter(bs, nil, nil), ShouldNotBeNil)
			So(NewComponentFilter(nil, bs, nil), ShouldNotBeNil)
			So(NewComponentFilter(nil, nil, bs), ShouldNotBeNil)

			So(NewComponentFilter(bs, bs, nil), ShouldNotBeNil)
			So(NewComponentFilter(nil, bs, bs), ShouldNotBeNil)
			So(NewComponentFilter(bs, nil, bs), ShouldNotBeNil)

			So(NewComponentFilter(bs, bs, bs), ShouldNotBeNil)
			So(NewComponentFilter(nil, nil, nil), ShouldNotBeNil)
		})
	})
}
