package main

import (
	"os"

	"github.com/jmattheis/goverter/cli"
)

func main() {
	cli.Run(os.Args, cli.RunOpts{})
}
