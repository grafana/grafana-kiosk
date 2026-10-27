ARG BUILD_PLATFORM=linux/amd64
ARG TARGET_PLATFORM=linux/amd64

FROM --platform=${BUILD_PLATFORM} golang:1.27.3-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS builder
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
