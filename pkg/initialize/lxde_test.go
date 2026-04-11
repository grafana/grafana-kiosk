package initialize

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSanitize(t *testing.T) {
	Convey("Given a string to sanitize", t, func() {
		Convey("When string contains newlines", func() {
			So(sanitize("hello\nworld"), ShouldEqual, "helloworld")
		})

		Convey("When string contains carriage returns", func() {
			So(sanitize("hello\rworld"), ShouldEqual, "helloworld")
		})

		Convey("When string contains both", func() {
			So(sanitize("a\r\nb\nc\r"), ShouldEqual, "abc")
		})

		Convey("When string is clean", func() {
			So(sanitize("clean"), ShouldEqual, "clean")
		})

		Convey("When string is empty", func() {
			So(sanitize(""), ShouldEqual, "")
		})
	})
}

func TestAllowedCommands(t *testing.T) {
	Convey("Given the allowed commands list", t, func() {
		Convey("Should allow lxpanel", func() {
			So(allowedCommands["/usr/bin/lxpanel"], ShouldBeTrue)
		})

		Convey("Should allow pcmanfm", func() {
			So(allowedCommands["/usr/bin/pcmanfm"], ShouldBeTrue)
		})

		Convey("Should allow xset", func() {
			So(allowedCommands["/usr/bin/xset"], ShouldBeTrue)
		})

		Convey("Should allow unclutter", func() {
			So(allowedCommands["/usr/bin/unclutter"], ShouldBeTrue)
		})

		Convey("Should reject arbitrary commands", func() {
			So(allowedCommands["/usr/bin/bash"], ShouldBeFalse)
			So(allowedCommands["/bin/sh"], ShouldBeFalse)
			So(allowedCommands["rm"], ShouldBeFalse)
		})

		Convey("Should reject path traversal attempts", func() {
			So(allowedCommands["/usr/bin/../bin/xset"], ShouldBeFalse)
		})
	})
}

func TestRunCommandAllowlist(t *testing.T) {
	Convey("Given the runCommand function", t, func() {
		Convey("When command is not in the allowlist", func() {
			// runCommand should return early without panicking
			// for disallowed commands
			So(func() {
				runCommand("/tmp", "/usr/bin/bash", []string{"-c", "echo hello"}, false)
			}, ShouldNotPanic)
		})

		Convey("When command is empty", func() {
			So(func() {
				runCommand("/tmp", "", []string{"arg"}, false)
			}, ShouldNotPanic)
		})
	})
}
