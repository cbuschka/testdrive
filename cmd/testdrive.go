package main

import (
	"fmt"
	"github.com/cbuschka/testdrive/internal"
	"os"
)

func main() {
	exitCode, err := internal.Run()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	os.Exit(exitCode)
}
