package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zhangqiwei/cxport/internal/tui"
)

var version = "dev"

func main() {
	format := flag.String("format", "", "Output format: xml (default) or md")
	ver := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *ver {
		fmt.Printf("cxport %s\n", version)
		os.Exit(0)
	}

	f := ""
	if *format != "" {
		f = *format
	}

	if err := tui.Run(f); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
