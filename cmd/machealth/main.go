package main

import (
	"fmt"
	"os"

	"github.com/lu-zhengda/machealth/internal/cli"
)

func main() {
	code, err := cli.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
