package main

import "os"

type Aa struct {
	I int
}

func main() {
	var a = Aa{}
	print(a.I)
	os.Exit(0)
}
