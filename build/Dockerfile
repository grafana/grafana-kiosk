ARG BUILD_PLATFORM=linux/amd64
ARG TARGET_PLATFORM=linux/amd64

FROM --platform=${BUILD_PLATFORM} golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder
ARG COMPILE_GOOS=linux
ARG COMPILE_GOARCH=amd64
ARG COMPILE_GOARM=""
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=${COMPILE_GOOS} GOARCH=${COMPILE_GOARCH} GOARM=${COMPILE_GOARM} go build -ldflags '-extldflags "-static"' -o grafana-kiosk ./pkg/cmd/grafana-kiosk/main.go
RUN mv /build/grafana-kiosk /

FROM --platform=${TARGET_PLATFORM} dtcooper/raspberrypi-os:latest@sha256:bc3c17c84cb4f20b17c3c776866115d9fa76c99702edb16b9c9c088458676145
RUN apt-get update && \
  apt-get install -qy \
  tzdata ca-certificates chromium-browser

WORKDIR /
COPY --from=builder /grafana-kiosk /kiosk/grafana-kiosk
ENTRYPOINT [ "/kiosk/grafana-kiosk" ]
