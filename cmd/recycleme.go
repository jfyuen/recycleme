package main

import (
	"flag"
	"fmt"
	"github.com/jfyuen/recycleme"
	"log"
	"os"
	"path"
)

var jsonFlag bool
var dirFlag string

func init() {
	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s -d DIR [options] EAN:\n", name)
		flag.PrintDefaults()
	}
	flag.BoolVar(&jsonFlag, "json", false, "Print json export")
	flag.StringVar(&dirFlag, "d", "", "Directory where to load product and packaging data")
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 || dirFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	recycleme.LoadJsonFiles(dirFlag, logger)

	product, err := recycleme.Scrap(flag.Arg(0))
	if err != nil {
		logger.Fatalln(err)
	}

	pkg := recycleme.NewProductPackage(*product)
	if jsonFlag {
		jsonBytes, err := pkg.ThrowAwayJson()
		if err != nil {
			logger.Fatalln(err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Println(pkg.ThrowAway())
	}
}
