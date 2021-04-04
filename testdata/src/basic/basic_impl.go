package basic

import (
	"context"
	"fmt"
)

type Handler struct {
	UnimplementedGreeterServer
}

func (h *Handler) SayHello(ctx context.Context, req *HelloRequest) *HelloReply {
	fmt.Println(req.Person.Name) // want "possible nil *Person use GetPerson function"
	return &HelloReply{
		Message: "Hello " + req.Person.GetName(),
	}
}
