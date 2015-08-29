package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	print("Entering\n")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	select {
	case sig := <-ch:
		fmt.Printf("Caught %d\n", sig)
	}

	print("Exiting\n")
}
