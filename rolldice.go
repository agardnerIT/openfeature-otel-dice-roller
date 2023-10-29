package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"github.com/open-feature/go-sdk-contrib/providers/flagd/pkg"
   	"github.com/open-feature/go-sdk/pkg/openfeature"
	"context"
	"time"
)

var (
	tracer  = otel.Tracer("rolldice")
	meter   = otel.Meter("rolldice")
	rollCnt metric.Int64Counter
	openFeatureClient = openfeature.NewClient("app")
)

func init() {
	var err error
	rollCnt, err = meter.Int64Counter("dice.rolls",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{roll}"))
	if err != nil {
		panic(err)
	}

	/* OpenFeature Initialisation
	 * Connect to feature flag system (flagd) here
	 */
	openfeature.SetProvider(flagd.NewProvider(
        flagd.WithHost("localhost"),
        flagd.WithPort(8013),
    ))
}

func rolldice(w http.ResponseWriter, r *http.Request) {

	feature_flag_key := "slow-your-roll"

	// Get flag value...
	// Evaluate your feature flag
    slowYourRoll, _ := openFeatureClient.BooleanValue(
        context.Background(), feature_flag_key, false, openfeature.EvaluationContext{},
    )

	ctx, span := tracer.Start(r.Context(), "roll")
	defer span.End()

	// Add feature flag values
	span.SetAttributes(attribute.String("feature_flag.key", feature_flag_key))
	span.SetAttributes(attribute.String("feature_flag.provider_name", "flagd"))
	span.SetAttributes(attribute.Bool("feature_flag.variant", slowYourRoll))

	if slowYourRoll {
		time.Sleep(2 * time.Second)
	}

	roll := 1 + rand.Intn(6)

	// Add the custom attribute to the span and counter.
	rollValueAttr := attribute.Int("roll.value", roll)
	span.SetAttributes(rollValueAttr)
	rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	resp := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}
