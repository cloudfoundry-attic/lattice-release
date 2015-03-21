package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	for {
		fmt.Fprintf(os.Stdout, "Hi from stdout. My args are: %s\n", os.Args[1:])
		fmt.Fprintln(os.Stderr, "Oopsie from stderr")
		time.Sleep(100 * time.Millisecond)
	}
}
