module github.com/grafana/grafana-kiosk

go 1.16

require (
	github.com/chromedp/cdproto v0.0.0-20210625233425-810000e4a4fc
	github.com/chromedp/chromedp v0.7.3
	github.com/grafana/grafana-api-golang-client v0.12.0
	github.com/ilyakaznacheev/cleanenv v1.2.5
	github.com/smartystreets/goconvey v1.6.4
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
)

replace github.com/grafana/grafana-api-golang-client v0.12.0 => github.com/nissessenap/grafana-api-golang-client v0.0.0-20221012135911-271ce27883ab
