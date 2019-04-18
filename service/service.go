package service

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"log"
	"time"
)

type Args struct {
	A, B   int
	Tracer map[string]string
}

type Reply struct {
	C int
}

type Arith struct {
	Tracer opentracing.Tracer
}

func (t *Arith) Add(args Args, reply *Reply) error {
	log.Printf("[Add] recieve: %#v", args)
	if args.Tracer != nil {
		wireContext, err := t.Tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(args.Tracer))
		if err == nil {
			span := t.Tracer.StartSpan("Add", ext.RPCServerOption(wireContext))
			defer span.Finish()
		} else {
			log.Printf("SpanMapTpContext: %v\n", err)
		}
	}
	time.Sleep(time.Second)
	reply.C = args.A + args.B
	return nil
}
