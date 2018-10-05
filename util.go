package main

import (
	"fmt"
	"os"
)

func logError(msg string, err error) {
	fmt.Fprintf(os.Stderr, msg+": %s\n", err) //nolint
}

func safeClose(c chan struct{}) {
	select {
	case <-c:
	default:
		close(c)
	}
}
