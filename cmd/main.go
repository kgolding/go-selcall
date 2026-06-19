package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"selcall"
)

func main() {
	mode := flag.String("a", "ZVEI1", "selcall mode: ZVEI1, CCIR, ZVEI2, EEA")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: selcall [-a MODE] VALUE OUTPUT\n  OUTPUT: path to .wav file, or - to write raw PCM to stdout")
		os.Exit(1)
	}
	value := flag.Arg(0)
	output := flag.Arg(1)

	var sc selcall.Selcall
	switch strings.ToUpper(*mode) {
	case "ZVEI1":
		sc = selcall.NewZVEI1()
	case "CCIR":
		sc = selcall.NewCCIR()
	case "ZVEI2":
		sc = selcall.NewZVEI2()
	case "EEA":
		sc = selcall.NewEEA()
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q: must be ZVEI1, CCIR, ZVEI2 or EEA\n", *mode)
		os.Exit(1)
	}

	if output == "-" {
		if err := sc.Write(value, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	f, err := os.Create(output)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := sc.WriteWav(value, f); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
