package main

import (
	"fmt"
	"github.com/cbuschka/testdrive/internal/cli"
	"os"
)

func main() {
	exitCode, err := cli.Run()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	os.Exit(exitCode)
}
