package main

import (
	"fmt"
	"os"
	"time"
)

var (
	logInfof = logInfofFunc()
)

func logInfofFunc() func(string, ...interface{}) {
	startup := time.Now()

	return func(format string, a ...interface{}) {
		d := time.Since(startup)
		fmt.Fprintf(os.Stdout, "%9d ", int(d.Seconds())) //nolint
		fmt.Fprintf(os.Stdout, format, a...)             //nolint
	}
}

func logError(msg string, err error) {
	fmt.Fprint(os.Stderr, msg)            //nolint
	fmt.Fprintf(os.Stderr, ": %s\n", err) //nolint
}

func logAcsError(msg string, err error) {
	msg = "while accessing http endpoint: " + msg
	logError(msg, err)
}

func safeClose(c chan struct{}) {
	select {
	case <-c:
	default:
		close(c)
	}
}
