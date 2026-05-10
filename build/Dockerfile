ARG BUILD_PLATFORM=linux/amd64
ARG TARGET_PLATFORM=linux/amd64

FROM --platform=${BUILD_PLATFORM} golang:1.26.3-alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d AS builder
ARG COMPILE_GOOS=linux
ARG COMPILE_GOARCH=amd64
ARG COMPILE_GOARM=""
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=${COMPILE_GOOS} GOARCH=${COMPILE_GOARCH} GOARM=${COMPILE_GOARM} go build -ldflags '-extldflags "-static"' -o grafana-kiosk ./pkg/cmd/grafana-kiosk/main.go
RUN mv /build/grafana-kiosk /

FROM --platform=${TARGET_PLATFORM} dtcooper/raspberrypi-os:latest@sha256:e44adffb34b1938218eb17f681e9e4c487523239bd07252e2199d403fd054eaf
RUN apt-get update && \
  apt-get install -qy \
  tzdata ca-certificates chromium-browser

WORKDIR /
COPY --from=builder /grafana-kiosk /kiosk/grafana-kiosk
ENTRYPOINT [ "/kiosk/grafana-kiosk" ]
