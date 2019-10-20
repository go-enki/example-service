package grpc

import (
	"context"

	example "github.com/go-enki/enki-example/proto"
)

func (ex ExampleService) Hello(ctx context.Context, req *example.HelloRequest) (res *example.HelloResponse, err error) {
	greeting, err := ex.greeting.Hello(ctx, req.Name)
	if err != nil {
		return nil, ErrorToProto(err)
	}
	return &example.HelloResponse{
		Greeting: GreetingToProto(greeting),
	}, nil
}
