package main

import (
	"fmt"
	"github.com/liyue201/opentracing-go-rpc-example/service"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"net"
	"net/rpc"
)

const (
	// Our service name.
	serviceName = "Server"

	// Host + port of our service.
	hostPort = "0.0.0.0.0"

	// Endpoint to send Zipkin spans to.
	zipkinHTTPEndpoint = "http://127.0.0.1:9411/api/v1/spans"

	// Debug mode.
	debug = false

	// same span can be set to true for RPC style spans (Zipkin V1) vs Node style (OpenTracing)
	sameSpan = true

	// make Tracer generate 128 bit traceID's for root spans.
	traceID128Bit = true
)

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

func run(tracer opentracing.Tracer) {
	arith := &service.Arith{
		tracer,
	}
	server := rpc.NewServer()
	server.Register(arith)

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":6234")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Errorf("Accept: %v", err)
			continue
		}
		go server.ServeConn(conn)
	}
}

func main() {
	collector, tracer := newTracer()
	run(tracer)
	collector.Close()
}
