// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	fmt "fmt"
	"io"
	"net"
	"os"

	"github.com/dingyalin/pinpoint-go-agent/integrations/nrgrpc"
	sampleapp "github.com/dingyalin/pinpoint-go-agent/integrations/nrgrpc/example/sampleapp"
	pinpoint "github.com/dingyalin/pinpoint-go-agent/pinpoint"
	"google.golang.org/grpc"
)

// Server is a gRPC server.
type Server struct{}

// processMessage processes each incoming Message.
func processMessage(ctx context.Context, msg *sampleapp.Message) {
	defer pinpoint.FromContext(ctx).StartSegment("processMessage").End()
	fmt.Printf("Message received: %s\n", msg.Text)
}

// DoUnaryUnary is a unary request, unary response method.
func (s *Server) DoUnaryUnary(ctx context.Context, msg *sampleapp.Message) (*sampleapp.Message, error) {
	processMessage(ctx, msg)
	return &sampleapp.Message{Text: "Hello from DoUnaryUnary"}, nil
}

// DoUnaryStream is a unary request, stream response method.
func (s *Server) DoUnaryStream(msg *sampleapp.Message, stream sampleapp.SampleApplication_DoUnaryStreamServer) error {
	processMessage(stream.Context(), msg)
	for i := 0; i < 3; i++ {
		if err := stream.Send(&sampleapp.Message{Text: "Hello from DoUnaryStream"}); nil != err {
			return err
		}
	}
	return nil
}

// DoStreamUnary is a stream request, unary response method.
func (s *Server) DoStreamUnary(stream sampleapp.SampleApplication_DoStreamUnaryServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&sampleapp.Message{Text: "Hello from DoStreamUnary"})
		} else if nil != err {
			return err
		}
		processMessage(stream.Context(), msg)
	}
}

// DoStreamStream is a stream request, stream response method.
func (s *Server) DoStreamStream(stream sampleapp.SampleApplication_DoStreamStreamServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if nil != err {
			return err
		}
		processMessage(stream.Context(), msg)
		if err := stream.Send(&sampleapp.Message{Text: "Hello from DoStreamStream"}); nil != err {
			return err
		}
	}
}

func main() {
	app, err := pinpoint.NewApplication(
		pinpoint.ConfigAppName("GoGrpcDemo"),
		pinpoint.ConfigAgentID("GoGrpcDemo"),
		pinpoint.ConfigCollectorUploaded(false),
		pinpoint.ConfigCollectorIP("127.0.0.1"),
		pinpoint.ConfigCollectorTCPPort(9994),
		pinpoint.ConfigCollectorStatPort(9995),
		pinpoint.ConfigCollectorSpanPort(9996),
		pinpoint.ConfigDebugLogger(os.Stdout),
	)
	if nil != err {
		panic(err)
	}

	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer(
		// Add the New Relic gRPC server instrumentation
		grpc.UnaryInterceptor(nrgrpc.UnaryServerInterceptor(app)),
		grpc.StreamInterceptor(nrgrpc.StreamServerInterceptor(app)),
	)
	sampleapp.RegisterSampleApplicationServer(grpcServer, &Server{})
	grpcServer.Serve(lis)
}