package main

import (
	"fmt"
	"os"

	"github.com/zkhvan/tfc/internal/build"
	"github.com/zkhvan/tfc/pkg/factory"
	"github.com/zkhvan/tfc/pkg/signal"
)

func main() {
	buildDate := build.Date
	buildVersion := build.Version

	f, err := factory.New(buildVersion)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := NewCmdRoot(f, buildVersion, buildDate).ExecuteContextC(signal.Notify()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
