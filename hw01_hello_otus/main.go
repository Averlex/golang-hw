package main

import (
	"fmt"

	rvrs "golang.org/x/example/hello/reverse"
)

// Reverse is a local wrapper over suggested reverse package.
func Reverse(str string) string {
	return rvrs.String(str)
}

func main() {
	fmt.Println(Reverse("Hello, OTUS!"))
}
