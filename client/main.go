package main

import (
	"context"
	"fmt"
	"github.com/liyue201/opentracing-go-rpc-example/service"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"log"
	"net/rpc"
	"time"
)

const (
	// Our service name.
	serviceName = "cli"

	// Host + port of our service.
	hostPort = "0.0.0.0.0"

	// Endpoint to send Zipkin spans to.
	zipkinHTTPEndpoint = "http://127.0.0.1:9411/api/v1/spans"

	// Debug mode.
	debug = false

	// Base endpoint of our SVC1 service.
	svc1Endpoint = "127.0.0.1:6234"

	// same span can be set to true for RPC style spans (Zipkin V1) vs Node style (OpenTracing)
	sameSpan = true

	// make Tracer generate 128 bit traceID's for root spans.
	traceID128Bit = true
)

func spanMap(tracer opentracing.Tracer, ctx context.Context) map[string]string {
	m := map[string]string{}
	if span := opentracing.SpanFromContext(ctx); span != nil {
		ext.SpanKindRPCClient.Set(span)
		if err := tracer.Inject(
			span.Context(),
			opentracing.TextMap,
			opentracing.TextMapCarrier(m),
		); err != nil {
			fmt.Printf("error encountered while trying to inject span: %+v\n", err)
		}
	}
	return m
}

func newTracer() (zipkin.Collector, opentracing.Tracer) {
	collector, err := zipkin.NewHTTPCollector(zipkinHTTPEndpoint)
	if err != nil {
		fmt.Printf("unable to create Zipkin HTTP collector: %+v\n", err)
		return nil, nil
	}

	// Create our recorder.
	recorder := zipkin.NewRecorder(collector, debug, hostPort, serviceName)

	// Create our tracer.
	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(sameSpan),
		zipkin.TraceID128Bit(traceID128Bit),
	)
	if err != nil {
		fmt.Printf("unable to create Zipkin tracer: %+v\n", err)
		return nil, nil
	}

	// Explicitly set our tracer to be the default tracer.
	opentracing.InitGlobalTracer(tracer)
	return collector, tracer
}

func callAdd(tracer opentracing.Tracer) {
	span := opentracing.StartSpan("callAdd")
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	defer span.Finish()

	client, err := rpc.Dial("tcp", svc1Endpoint)
	if err != nil {
		span.SetTag("error", err.Error())
		log.Printf("dial error: %v\n", err.Error())
		return
	}
	defer client.Close()

	tracerM := spanMap(tracer, ctx)
	log.Printf("m = %#v", tracerM)

	req := service.Args{2, 3, tracerM}
	var reply service.Reply
	err = client.Call("Arith.Add", req, &reply)
	if err != nil {
		span.SetTag("error", err.Error())
		log.Printf("Call Arith.Add error: %v\n", err.Error())
		return
	}
	time.Sleep(time.Second / 2)
	log.Printf("Call Arith.Add reply: %#v\n", reply)
}

func main() {
	collector, tracer := newTracer()
	for i := 0; i < 2; i++ {
		callAdd(tracer)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Minute)
	collector.Close()
}
