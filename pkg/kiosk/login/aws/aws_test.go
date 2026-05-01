package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAwsLoginFlow(t *testing.T) {
	baseCfg := func() *config.Config {
		return &config.Config{
			General: config.General{Mode: "full", AutoFit: true, PageLoadDelayMS: 0},
			Target:  config.Target{URL: "https://grafana.example.com/d/abc", Username: "admin", Password: "secret"},
		}
	}

	Convey("Given awsLoginFlow", t, func() {
		mock := browsertest.NewMock()
		cfg := baseCfg()
		url := shared.GenerateURL(cfg)

		Convey("Returns error if Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := awsLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
		})

		Convey("Full sequence without MFA: navigate → cookie accept → SSO → credentials", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- awsLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, url)
			waitCalls := mock.CallsTo("WaitVisible")
			So(waitCalls[0].Args[0], ShouldContainSubstring, "login/sso")
			So(waitCalls[1].Args[0], ShouldContainSubstring, "awsccc-cb-buttons")
			So(mock.CallCount("Click"), ShouldEqual, 2)
			So(mock.CallCount("SendKeys"), ShouldEqual, 2)
			So(mock.CallCount("WaitNotVisible"), ShouldEqual, 0)
		})

		Convey("With MFA: calls WaitNotVisible after credentials", func() {
			cfg.Target.UseMFA = true
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- awsLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			So(mock.CallCount("WaitNotVisible"), ShouldEqual, 1)
			So(mock.CallsTo("WaitNotVisible")[0].Args[0], ShouldContainSubstring, "awsui-input-2")
		})

		Convey("Reloads on message", func() {
			ctx, cancel := context.WithCancel(context.Background())
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- awsLoginFlow(ctx, cfg, mock, url, messages) }()
			time.Sleep(10 * time.Millisecond)
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldEqual, 2)
		})

		Convey("Returns error if WaitNotVisible fails during MFA", func() {
			cfg.Target.UseMFA = true
			mock.Errors["WaitNotVisible"] = errors.New("timeout")
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- awsLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			err := <-done
			So(err, ShouldNotBeNil)
		})
	})
}
