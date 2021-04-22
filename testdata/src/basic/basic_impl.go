package basic

import (
	"context"
	"fmt"
)

type Handler struct {
	UnimplementedGreeterServer
}

func (h *Handler) SayHello(ctx context.Context, requ *HelloRequest) (*HelloReply, error) {
	fmt.Println((*requ).GetPerson())
	for requ.GetPerson().GetName() == "Narek" {
		fmt.Println(requ.Person.Name) // want "possible nil *Person use GetPerson function"
	}
	return nil, nil
}
