package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	port = flag.String("port", os.Getenv("PORT"), "port ")
)

func main() {
	fmt.Printf("Hello and welcome, %s!\n", *port)
}
