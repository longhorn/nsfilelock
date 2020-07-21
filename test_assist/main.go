package main

import (
	"fmt"
	"os"

	"github.com/longhorn/nsfilelock"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Usage: %s <file> <ns>\n", args[0])
		os.Exit(-1)
	}

	lockFile := args[1]

	ns := ""
	if len(args) > 2 {
		ns = args[2]
	}

	lock := nsfilelock.NewLock(ns, lockFile)
	if err := lock.Lock(); err != nil {
		fmt.Printf("Fail to lock %s in %s: %v\n", lockFile, ns, err)
		os.Exit(-1)
	}
	fmt.Println(nsfilelock.SuccessResponse)
}
