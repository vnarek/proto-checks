package main

func main() {
	var x *int = nil
	if 2+2 == 4 {
		y := new(int)
		var z **int = &y
		x = *z
	} else {
		x = new(int)
	}
	z := x
}
