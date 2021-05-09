package basic

import (
	"context"
	"fmt"
	"math/rand"
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

func (h *Handler) Rng() {
	def := new(int)
	x := def
	if *x == rand.Int() {
		x = nil
	} else {
		*x = 42
	}
	y := *x // want "x could be nil"
	fmt.Sprint(y)
}

func (h *Handler) NotNil() {
	var x *int
	if rand.Int() == 0 {
		x = new(int)
	} else {
		x = new(int)
	}
	fmt.Sprint(*x)
}

func (h *Handler) SwitchRng() {
	x := new(int)
	*x = rand.Int()
	switch *x {
	case 0:
		*x = *x + 1
	case 1:
		x = nil
	case 2:
		*x = *x - 1
	default:
		*x = *x / *x
	}
	*x = 0 // want "x could be nil"
}