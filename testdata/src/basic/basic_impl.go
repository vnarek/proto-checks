package basic

import (
	"context"
	"fmt"
)

type Handler struct {
	UnimplementedGreeterServer
}

func (h *Handler) SayHello(ctx context.Context, requ *HelloRequest) (*HelloReply, error) {
	req := *requ // want "requ could be nil"
	fmt.Sprint(req)
	return nil, nil
}

func (h *Handler) HiHello(ctx context.Context, a *int, b *bool) {
	c := a
	d := *c // want "c could be nil"

	fmt.Sprintln(c, d)
}
