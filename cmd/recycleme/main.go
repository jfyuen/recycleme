package main

import (
	"flag"
	"fmt"
	"github.com/jfyuen/recycleme"
	"log"
	"net/http"
	"os"
	"path"
)

var jsonFlag bool
var serverFlag bool
var dirFlag string
var serverPort string

func init() {
	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s -d DIR [options] EAN:\n", name)
		flag.PrintDefaults()
	}
	flag.BoolVar(&jsonFlag, "json", false, "Print json export")
	flag.BoolVar(&serverFlag, "server", false, "Run in server mode, serving json (EAN as input is useless)")
	flag.StringVar(&dirFlag, "d", "", "Directory where to load product and packaging data")
	flag.StringVar(&serverPort, "p", "8080", "Port to listen to")
}

func main() {
	flag.Parse()
	if (len(flag.Args()) != 1 && !serverFlag) || (serverFlag && len(flag.Args()) != 0) || dirFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	recycleme.LoadJSONFiles(dirFlag, logger)

	if serverFlag {
		http.HandleFunc("/bin/", recycleme.BinHandler)
		http.HandleFunc("/bins/", recycleme.BinsHandler)
		http.HandleFunc("/materials/", recycleme.MaterialsHandler)
		http.HandleFunc("/throwaway/", recycleme.ThrowAwayHandler)
		http.HandleFunc("/", recycleme.HomeHandler)
		fs := http.FileServer(http.Dir("data/static"))

		http.Handle("/static/", http.StripPrefix("/static/", fs))
		err := http.ListenAndServe(":"+serverPort, nil)
		if err != nil {
			logger.Fatalln(err)
		}
		logger.Println("Running in server mode")
	} else {
		product, err := recycleme.Scrap(flag.Arg(0))
		if err != nil {
			logger.Fatalln(err)
		}
		pkg := recycleme.NewProductPackage(product)
		if jsonFlag {
			jsonBytes, err := pkg.ThrowAwayJSON()
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Println(string(jsonBytes))
		} else {
			logger.Println(pkg.ThrowAway())
		}
	}
}
