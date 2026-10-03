package main

import (
	"os"

	"github.com/mvdatacenter/mvdata-cli/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
