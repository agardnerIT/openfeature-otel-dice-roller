package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"context"
	"time"

	flagd "github.com/open-feature/go-sdk-contrib/providers/flagd/pkg"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	tracer            = otel.Tracer("rolldice")
	meter             = otel.Meter("rolldice")
	rollCnt           metric.Int64Counter
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

// Redirect homepage to /rolldice
func renderHomepage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "rolldice", http.StatusPermanentRedirect)
}

func rolldice(w http.ResponseWriter, r *http.Request) {

	feature_flag_key := "slow-your-roll"

	// Pass UserAgent string to flagd
	evaluationContext := openfeature.NewEvaluationContext("", map[string]interface{}{
		"userAgent": r.UserAgent(),
	})

	// Get flag value...
	// Evaluate your feature flag
	slowYourRoll, _ := openFeatureClient.BooleanValue(
		context.Background(), feature_flag_key, false, evaluationContext,
	)

	ctx, span := tracer.Start(r.Context(), "roll")
	defer span.End()

	evaluationContextString := fmt.Sprintf("%#v", evaluationContext)

	// Add feature flag values
	span.SetAttributes(attribute.String("feature_flag.key", feature_flag_key))
	span.SetAttributes(attribute.String("feature_flag.provider_name", "flagd"))
	span.SetAttributes(attribute.Bool("feature_flag.variant", slowYourRoll))
	span.SetAttributes(attribute.String("feature_flag.evaluation_context", evaluationContextString))

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
