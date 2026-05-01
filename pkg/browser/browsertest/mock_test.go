package browsertest

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMock(t *testing.T) {
	Convey("Given a browser mock", t, func() {
		m := NewMock()
		ctx := context.Background()

		Convey("Should record Navigate calls", func() {
			err := m.Navigate(ctx, "https://example.com")
			So(err, ShouldBeNil)
			So(m.Calls, ShouldHaveLength, 1)
			So(m.Calls[0].Method, ShouldEqual, "Navigate")
			So(m.Calls[0].Args, ShouldResemble, []string{"https://example.com"})
		})

		Convey("Should record WaitVisible calls", func() {
			err := m.WaitVisible(ctx, `//input[@name="user"]`)
			So(err, ShouldBeNil)
			So(m.CallsTo("WaitVisible"), ShouldHaveLength, 1)
		})

		Convey("Should record SendKeys calls", func() {
			err := m.SendKeys(ctx, `//input[@name="user"]`, "admin")
			So(err, ShouldBeNil)
			So(m.Calls[0].Args, ShouldResemble, []string{`//input[@name="user"]`, "admin"})
		})

		Convey("Should return preconfigured errors", func() {
			m.Errors["Navigate"] = errors.New("connection refused")
			err := m.Navigate(ctx, "https://example.com")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "connection refused")
			So(m.Calls, ShouldHaveLength, 1)
		})

		Convey("Should count calls per method", func() {
			_ = m.Navigate(ctx, "https://a.com")
			_ = m.Navigate(ctx, "https://b.com")
			_ = m.WaitVisible(ctx, "sel")
			So(m.CallCount("Navigate"), ShouldEqual, 2)
			So(m.CallCount("WaitVisible"), ShouldEqual, 1)
			So(m.CallCount("Click"), ShouldEqual, 0)
		})

		Convey("Should reset recorded calls", func() {
			_ = m.Navigate(ctx, "https://example.com")
			m.Reset()
			So(m.Calls, ShouldHaveLength, 0)
		})
	})
}
