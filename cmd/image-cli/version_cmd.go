package main

import (
	"flag"
	"os"

	"github.com/neko233-com/image-cli/internal/version"
)

func runVersion(args []string) {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "JSON 输出")
	_ = fs.Parse(args)
	if err := version.Print(os.Stdout, *asJSON); err != nil {
		fail("%v", err)
	}
}
