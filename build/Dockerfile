ARG BUILD_PLATFORM=linux/amd64
ARG TARGET_PLATFORM=linux/amd64

FROM --platform=${BUILD_PLATFORM} golang:1.27.1-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder
ARG COMPILE_GOOS=linux
ARG COMPILE_GOARCH=amd64
ARG COMPILE_GOARM=""
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=${COMPILE_GOOS} GOARCH=${COMPILE_GOARCH} GOARM=${COMPILE_GOARM} go build -ldflags '-extldflags "-static"' -o grafana-kiosk ./pkg/cmd/grafana-kiosk/main.go
RUN mv /build/grafana-kiosk /

FROM --platform=${TARGET_PLATFORM} dtcooper/raspberrypi-os:latest@sha256:1b54403e05f3d868e439d72cca54706309376120183f12f551b98ac0eb13bc1c
RUN apt-get update && \
  apt-get install -qy \
  tzdata ca-certificates chromium-browser

WORKDIR /
COPY --from=builder /grafana-kiosk /kiosk/grafana-kiosk
ENTRYPOINT [ "/kiosk/grafana-kiosk" ]
