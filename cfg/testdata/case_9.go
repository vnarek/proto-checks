package main

func main() {
	x := new(int)
	max := new(int)
	*max = 5
	for i := *x; i < *max; i++ {
		x = &i
	}
	x = nil
}
