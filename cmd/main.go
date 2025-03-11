package main

import (
	"context"
	"os"

	"github.com/GBA-BI/tes-filer/cmd/filer"
)

func main() {
	command := filer.NewFilerCommand(context.Background())
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
