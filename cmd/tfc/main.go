package main

import (
	"fmt"
	"os"

	"github.com/zkhvan/tfc/pkg/factory"
	"github.com/zkhvan/tfc/pkg/signal"
)

func main() {
	f, err := factory.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := NewCmdRoot(f).ExecuteContextC(signal.Notify()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
