package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestKiosk checks kiosk command
func TestKioskCommand(t *testing.T) {
	SkipConvey("Kiosk args", t, func() {
		So(true, ShouldBeTrue)
	})
}
