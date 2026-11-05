package apikey

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIsDataSourceQueryRequest(t *testing.T) {
	Convey("Given a request URL and target host", t, func() {
		Convey("When request is a datasource query to the target host", func() {
			result := IsDataSourceQueryRequest(
				"https://grafana.example.com/api/ds/query?ds_type=prometheus",
				"https", "grafana.example.com",
			)
			So(result, ShouldBeTrue)
		})

		Convey("When request is a datasource query with port", func() {
			result := IsDataSourceQueryRequest(
				"https://localhost:3000/api/ds/query?ds_type=loki",
				"https", "localhost:3000",
			)
			So(result, ShouldBeTrue)
		})

		Convey("When request matches host but is not a query path", func() {
			result := IsDataSourceQueryRequest(
				"https://grafana.example.com/api/dashboards/home",
				"https", "grafana.example.com",
			)
			So(result, ShouldBeFalse)
		})

		Convey("When request is a query path but wrong host", func() {
			result := IsDataSourceQueryRequest(
				"https://other.example.com/api/ds/query?ds_type=prometheus",
				"https", "grafana.example.com",
			)
			So(result, ShouldBeFalse)
		})

		Convey("When request scheme does not match", func() {
			result := IsDataSourceQueryRequest(
				"http://grafana.example.com/api/ds/query?ds_type=prometheus",
				"https", "grafana.example.com",
			)
			So(result, ShouldBeFalse)
		})

		Convey("When request host is a prefix of target host", func() {
			result := IsDataSourceQueryRequest(
				"https://grafana.example.com.evil.com/api/ds/query?ds_type=prometheus",
				"https", "grafana.example.com",
			)
			So(result, ShouldBeFalse)
		})

		Convey("When request uses the Grafana Cloud query API path", func() {
			result := IsDataSourceQueryRequest(
				"https://myorg.grafana.net/apis/query.grafana.app/v0alpha1/namespaces/stacks-123/query?ds_type=prometheus&requestId=SQR1",
				"https", "myorg.grafana.net",
			)
			So(result, ShouldBeTrue)
		})

		Convey("When request uses the Grafana Cloud query API path with port", func() {
			result := IsDataSourceQueryRequest(
				"https://localhost:3000/apis/query.grafana.app/v0alpha1/namespaces/default/query?ds_type=loki",
				"https", "localhost:3000",
			)
			So(result, ShouldBeTrue)
		})

		Convey("When request uses Grafana Cloud API path but wrong host", func() {
			result := IsDataSourceQueryRequest(
				"https://other.grafana.net/apis/query.grafana.app/v0alpha1/namespaces/stacks-123/query?ds_type=prometheus",
				"https", "myorg.grafana.net",
			)
			So(result, ShouldBeFalse)
		})

		Convey("When request has /apis/ path but is not a query endpoint", func() {
			result := IsDataSourceQueryRequest(
				"https://grafana.example.com/apis/dashboard.grafana.app/v1/namespaces/default/dashboards",
				"https", "grafana.example.com",
			)
			So(result, ShouldBeFalse)
		})

		Convey("When request is under query.grafana.app but not a /query endpoint", func() {
			result := IsDataSourceQueryRequest(
				"https://myorg.grafana.net/apis/query.grafana.app/v0alpha1/namespaces/stacks-123/something-else",
				"https", "myorg.grafana.net",
			)
			So(result, ShouldBeFalse)
		})
	})
}

func TestIsTargetHostRequest(t *testing.T) {
	Convey("Given a request host and target host", t, func() {
		Convey("When hosts match exactly", func() {
			So(IsTargetHostRequest("grafana.example.com", "grafana.example.com"), ShouldBeTrue)
		})

		Convey("When hosts match with port", func() {
			So(IsTargetHostRequest("localhost:3000", "localhost:3000"), ShouldBeTrue)
		})

		Convey("When hosts do not match", func() {
			So(IsTargetHostRequest("other.example.com", "grafana.example.com"), ShouldBeFalse)
		})

		Convey("When request host is a subdomain of target", func() {
			So(IsTargetHostRequest("sub.grafana.example.com", "grafana.example.com"), ShouldBeFalse)
		})

		Convey("When ports differ", func() {
			So(IsTargetHostRequest("localhost:3001", "localhost:3000"), ShouldBeFalse)
		})
	})
}

func TestApikeyLoginFlow(t *testing.T) {
	Convey("Given apikeyLoginFlow", t, func() {
		mock := browsertest.NewMock()
		cfg := &config.Config{General: config.General{PageLoadDelayMS: 0}}
		url := "https://grafana.example.com/d/abc?kiosk=1"

		Convey("Navigates to provided URL", func() {
			mock.Errors["Navigate"] = errors.New("stop")
			_ = apikeyLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(mock.Calls[0].Args[0], ShouldEqual, url)
		})

		Convey("Returns error if Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := apikeyLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
		})

		Convey("Exits cleanly on context cancel", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- apikeyLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			cancel()
			So(<-done, ShouldBeNil)
		})

		Convey("Reloads on message", func() {
			ctx, cancel := context.WithCancel(context.Background())
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- apikeyLoginFlow(ctx, cfg, mock, url, messages) }()
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldEqual, 2)
		})
	})
}

func headerMap(entries []*fetch.HeaderEntry) map[string][]string {
	m := map[string][]string{}
	for _, e := range entries {
		m[e.Name] = append(m[e.Name], e.Value)
	}
	return m
}

func TestRequestHeaders(t *testing.T) {
	Convey("Given an intercepted request", t, func() {
		Convey("When it is a POST to the target host, the original headers are kept", func() {
			orig := network.Headers{
				"Accept":           "application/json, text/plain, */*",
				"Content-Type":     "application/json",
				"X-Grafana-Org-Id": "1",
			}
			h := headerMap(RequestHeaders(orig, false, true, "glsa_test"))
			So(h["Content-Type"], ShouldResemble, []string{"application/json"})
			So(h["Accept"], ShouldResemble, []string{"application/json, text/plain, */*"})
			So(h["X-Grafana-Org-Id"], ShouldResemble, []string{"1"})
			So(h["Authorization"], ShouldResemble, []string{"Bearer glsa_test"})
		})

		Convey("When it already has an Authorization header, it is replaced", func() {
			orig := network.Headers{"authorization": "Basic abc"}
			h := headerMap(RequestHeaders(orig, false, true, "glsa_test"))
			So(h, ShouldNotContainKey, "authorization")
			So(h["Authorization"], ShouldResemble, []string{"Bearer glsa_test"})
		})

		Convey("When it is a datasource query without Content-Type, JSON is added", func() {
			h := headerMap(RequestHeaders(network.Headers{"Accept": "*/*"}, true, true, "k"))
			So(h["Content-Type"], ShouldResemble, []string{"application/json"})
		})

		Convey("When it is a datasource query with Content-Type, it is not duplicated", func() {
			orig := network.Headers{"content-type": "application/json"}
			h := headerMap(RequestHeaders(orig, true, true, "k"))
			So(h["content-type"], ShouldResemble, []string{"application/json"})
			So(h, ShouldNotContainKey, "Content-Type")
		})
	})
}
