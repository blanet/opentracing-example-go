package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

func init() {
	// Sample configuration for testing. Use constant sampling to sample every trace
	// and enable LogSpan to log every span via configured Logger.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	_, err := cfg.InitGlobalTracer(
		"ts_server",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
	}
}

func handler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var err error
	span, ctx := opentracing.StartSpanFromContext(ctx, "begin")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	err = auth(ctx)
	if err != nil {
		return
	}
	err = query(ctx)

	fmt.Fprintf(res, "hello %d\n", time.Now().Unix())
}

func authInner(ctx context.Context) error {
	var err error

	span, ctx := opentracing.StartSpanFromContext(ctx, "auth_inner")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	select {
	case <-time.After(time.Duration(rand.Int63n(1100)) * time.Millisecond):
		return nil
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}

func auth(ctx context.Context) error {
	var err error

	span, ctx := opentracing.StartSpanFromContext(ctx, "auth")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	err = authInner(ctx)
	if err != nil {
		return err
	}

	select {
	case <-time.After(time.Duration(rand.Int63n(1100)) * time.Millisecond):
		return nil
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}

func query(ctx context.Context) error {
	var err error

	span, ctx := opentracing.StartSpanFromContext(ctx, "query_db")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	select {
	case <-time.After(time.Duration(rand.Int63n(1100)) * time.Millisecond):
		return nil
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":9999", mux))
}
