package basic

import (
	"context"
	"fmt"
)

type Handler struct {
	UnimplementedGreeterServer
}

func (h *Handler) SayHello(ctx context.Context, requ *HelloRequest) (*HelloReply, error) {
	fmt.Println(requ.Person.Name) // want "possible nil *Person use GetPerson function"
	return &HelloReply{
		Message: "Hello " + requ.Person.GetName(),
	}, nil
}
