#!/bin/sh

JAEGER_VERSION=1.50
FLAGD_VERSION=latest
OTEL_COLLECTOR_ADDR=localhost
OTEL_COLLECTOR_PORT=4317

## Start Jaeger
echo ">> Starting Jaeger and OpenTelemetry Collector"
docker run \
  --rm -d \
  --name jaeger \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 16686:16686 \
  jaegertracing/all-in-one:${JAEGER_VERSION}

## Start flagd
echo ">> Starting flagd"
docker run \
  --rm -d \
  --name flagd \
  -p 8013:8013 \
  -v $(pwd):/etc/flagd \
  ghcr.io/open-feature/flagd:${FLAGD_VERSION} start \
  --uri file:./etc/flagd/demo.flagd.json \
  --metrics-exporter otel \
  --otel-collector-uri ${OTEL_COLLECTOR_ADDR}:${OTEL_COLLECTOR_PORT}

go mod tidy
