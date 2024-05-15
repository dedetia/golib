package otel

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"runtime"
	"strings"
)

var (
	appName string
	tracer  trace.Tracer
)

type Options struct {
	AppName          string
	JaegerURL        string
	TracerSampleRate float64
}

func init() {
	tracer = otel.Tracer(appName)
}

func (o *Options) init() {
	if o.AppName == "" {
		o.AppName = "tracer-app"
	}

	if o.TracerSampleRate <= 0 {
		o.TracerSampleRate = 1
	}

	appName = o.AppName
}

func NewTracer(opt *Options) (trace.TracerProvider, error) {
	opt.init()

	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpointURL(opt.JaegerURL),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(opt.TracerSampleRate)),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(opt.AppName),
			attribute.Int64("ID", 1),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func Tracer(ctx context.Context) (context.Context, trace.Span) {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	funcNameSplit := strings.SplitN(funcName, ".", 2)
	replacer := strings.NewReplacer("(", "", ")", "", "*", "")
	operation := replacer.Replace(funcNameSplit[1])

	return tracer.Start(ctx, operation)
}
