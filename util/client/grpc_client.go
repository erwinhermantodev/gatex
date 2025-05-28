package util

import (
	"context"
	"log"
	"time"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial: grpc server with new relic or elastic apm middleware,
func Dial(addr string, opts ...grpc.UnaryClientInterceptor) *grpc.ClientConn {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcMiddleware.ChainUnaryClient(opts...)),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(25*1024*1024), // Set to 10MB
			grpc.MaxCallSendMsgSize(25*1024*1024), // Set to 10MB
		),
	)

	if err != nil {
		log.Fatal("could not connect to", addr, err)
	}
	return conn
}
