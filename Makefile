CCARMV7=arm-linux-gnueabihf-gcc
CCARM64=aarch64-linux-gnu-gcc

all: grafana-kiosk
	@echo "Building"

dev:
	@echo "Building grafana-kiosk"
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -o bin/grafana-kiosk.darwin pkg/cmd/grafana-kiosk/main.go

grafana-kiosk: dev
	@echo "Building grafana-kiosk"
	mkdir -p bin
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/grafana-kiosk.linux.amd64 pkg/cmd/grafana-kiosk/main.go
	GO111MODULE=on GOOS=linux GOARCH=386 go build -o bin/grafana-kiosk.linux.386 pkg/cmd/grafana-kiosk/main.go
	GO111MODULE=on GOOS=linux GOARCH=arm GOARM=5 go build -o bin/grafana-kiosk.linux.armv5 pkg/cmd/grafana-kiosk/main.go
	GO111MODULE=on GOOS=linux GOARCH=arm GOARM=6 go build -o bin/grafana-kiosk.linux.armv6 pkg/cmd/grafana-kiosk/main.go
	GO111MODULE=on GOOS=linux GOARCH=arm GOARM=7 go build -o bin/grafana-kiosk.linux.armv7 pkg/cmd/grafana-kiosk/main.go
	GO111MODULE=on GOOS=linux GOARCH=arm64 go build -o bin/grafana-kiosk.linux.arm64 pkg/cmd/grafana-kiosk/main.go
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -trimpath -o bin/grafana-kiosk.darwin.amd64 -a -tags netgo -ldflags '-s -w' pkg/cmd/grafana-kiosk/main.go
	GO111MODULE=on GOOS=windows GOARCH=amd64 go build -o bin/grafana-kiosk.windows.amd64.exe -a -tags netgo -ldflags '-w' pkg/cmd/grafana-kiosk/main.go

circleci-lint:
	@echo "Linting in circleci"
	circleci local execute --job cmd_lint

circleci-test:
	@echo "Testing in circleci"
	circleci local execute --job cmd_test

circleci-build:
	@echo "Build in circleci"
	circleci local execute --job build

package: grafana-kiosk
	@echo "Packaging"

release: package
	@echo "Release"
