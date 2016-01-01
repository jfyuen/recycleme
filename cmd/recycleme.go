package main

import (
	"fmt"
	"os"
	"flag"
	"path"
	"github.com/jfyuen/recycleme"
)

var jsonFlag bool
func init() {
	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [options] EAN:\n", name)
		flag.PrintDefaults()
	}
	flag.BoolVar(&jsonFlag, "json", false, "Print json export")
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	product, err := recycleme.Scrap(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if jsonFlag {
		jsonBytes, err := product.Json()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Println(product)
	}
}
