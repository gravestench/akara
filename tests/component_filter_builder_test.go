package tests

import (
	"github.com/gravestench/akara"
	"testing"

	"github.com/gravestench/bitset"

	. "github.com/smartystreets/goconvey/convey"
)

func TestComponentFilterBuilder(t *testing.T) {
	Convey("Given an Arbitrary BitSet", t, func() {
		bs := bitset.NewBitSet()

		Convey("A component filter may be created with any combination of requirements/restrictions", func() {
			So(akara.NewComponentFilter(bs, nil, nil), ShouldNotBeNil)
			So(akara.NewComponentFilter(nil, bs, nil), ShouldNotBeNil)
			So(akara.NewComponentFilter(nil, nil, bs), ShouldNotBeNil)

			So(akara.NewComponentFilter(bs, bs, nil), ShouldNotBeNil)
			So(akara.NewComponentFilter(nil, bs, bs), ShouldNotBeNil)
			So(akara.NewComponentFilter(bs, nil, bs), ShouldNotBeNil)

			So(akara.NewComponentFilter(bs, bs, bs), ShouldNotBeNil)
			So(akara.NewComponentFilter(nil, nil, nil), ShouldNotBeNil)
		})
	})
}
