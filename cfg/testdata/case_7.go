package main

func main() {
	x := new(5)

	if y := x; x == 5 {
		*x = **z
	} else {
		**x = &z
	}
}
