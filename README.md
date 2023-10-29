# OpenFeature & OpenTelemetry Dice Roller

![](assets/apptrace.jpg)

The OpenTelemetry dice roll example instrumented with [OpenFeature](https://openfeature.dev), [Jaeger](https://jaegertracing.io) and [flagd](https://flagd.dev).

This is the OpenTelemetry "getting started" code, slightly modified to:

1) Send traces to Jaeger and not stdout
2) The dice roller application is feature flagged to add an artificial pause
3) The dice roller application adds the feature flag information to the span data according to the [OpenTelemetry specifications for feature flags](https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-spans/).

In this setup, both flagd (the flag backend system) **and** the "dice roller" application will both generate OpenTelemetry traces.

This provides two "lenses" on the data:

1) From the perspective of the flag backend system operator (ie. How healthy is my flag backend)
2) From the user's perspective (ie. What experience are my user's receiving when using the system


# Run in the Cloud
TODO

---------------------

# Run Locally

## Step 1: Start Jaeger

Start [Jaeger](https://jaegertracing.io) so there is a trace backend available to send traces to:

```
docker run --rm --name jaeger \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 16686:16686 \
  jaegertracing/all-in-one:1.50
```

## Step 2: Start flagd

Start flagd and have it send traces to Jaeger:

```
docker run \
  --rm -it \
  --name flagd \
  -p 8013:8013 \
  -v $(pwd):/etc/flagd \
  ghcr.io/open-feature/flagd:latest start \
  --uri file:./etc/flagd/demo.flagd.json \
  --metrics-exporter otel \
  --otel-collector-uri localhost:4317
```

or (if you have flagd in your `PATH`):
```
flagd start --port 8013 --uri file:./demo.flagd.json --metrics-exporter otel --otel-collector-uri localhost:4317
```

## Step 3: Run application

```
go mod tidy
go run .
```

Access the application: `http://localhost:8080/rolldice`

## Step 4: View Application Traces

View traces in Jaeger: `http://localhost:16686/search?service=dice-roller`

![](assets/apptrace.jpg)

## Step 5: View flagd Traces

View traces in Jaeger: `http://localhost:16686/search?service=flagd`

![](assets/jaeger-flagd-traces.jpg)

## Step 6: Change flag definition

It is time to slow your roll. Do this by changing the feature flag definition (effectively, turning it on).

Modify [line 9 of demo.flagd.json](https://github.com/agardnerIT/openfeature-otel-dice-roller/blob/45b8496620cfed77c54a21f8526661c9e31b9cc6/demo.flagd.json#L9) and change `defaultValue` from `off` to `on`.

This will make flagd return the `on` variant with the value `true`.

> Note: No restarts are necessary. Flagd is watching this file so will automatically pick up the changes and emit an OpenTelemetry trace

In Jaeger under the `flagd` service, notice there is a `flagSync` trace with the relevant information including how many flags were updated.

![](assets/flagsync.jpg)


