package cache

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// NewInterceptor creates an interceptor for use with gRPC.
func NewInterceptor(registry Registry) grpc.UnaryServerInterceptor {
	return interceptor{registry: registry}.unaryServer
}

type interceptor struct {
	registry Registry
}

func (i interceptor) unaryServer(
	ctx context.Context,
	request interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	stringer, ok := request.(fmt.Stringer)
	if !ok {
		return handler(ctx, request)
	}

	cache, ok := i.registry.GetByName(info.FullMethod)
	if !ok {
		return handler(ctx, request)
	}

	key := stringer.String()

	value, ok := cache.Get(key)
	if ok {
		return value, nil
	}

	response, err := handler(ctx, request)
	if err == nil {
		cache.Put(key, response)
	}

	return response, err
}
