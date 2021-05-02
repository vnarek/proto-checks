package main

func main() {
	x := new(int)
	max := new(int)
	*max = 5
	for i := *x; i < *max; i++ {
		{}
		{
			for j := *x; j < *max; j++ {
				{
					a = b
				}
			}
		}
		{}
	}
	x = nil
}