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

var jsonFlag = flag.Bool("json", false, "Print json export")
var serverFlag = flag.Bool("server", false, "Run in server mode, serving json (EAN as input is useless)")
var dirFlag = flag.String("d", "", "Directory where to load product and packaging data")
var serverPort = flag.String("p", "8080", "Port to listen to")

func init() {
	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s -d DIR [options] EAN:\n", name)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if (len(flag.Args()) != 1 && !*serverFlag) || (*serverFlag && len(flag.Args()) != 0) || *dirFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	recycleme.LoadJSONFiles(*dirFlag, logger)
	fetcher, err := recycleme.NewDefaultFetcher()
	if err != nil {
		logger.Println(err.Error())
	}
	if *serverFlag {
		http.HandleFunc("/bin/", recycleme.BinHandler)
		http.HandleFunc("/bins/", recycleme.BinsHandler)
		http.HandleFunc("/materials/", recycleme.MaterialsHandler)
		http.HandleFunc("/blacklist/add/", func(w http.ResponseWriter, r *http.Request) {
			recycleme.Blacklist.AddBlacklistHandler(w, r, logger, fetcher)
		})
		http.HandleFunc("/throwaway/", func(w http.ResponseWriter, r *http.Request) {
			recycleme.ThrowAwayHandler(w, r, fetcher)
		})
		http.HandleFunc("/", recycleme.HomeHandler)
		fs := http.FileServer(http.Dir("data/static"))

		http.Handle("/static/", http.StripPrefix("/static/", fs))
		err := http.ListenAndServe(":"+*serverPort, nil)
		if err != nil {
			logger.Fatalln(err)
		}
		logger.Println("Running in server mode")
	} else {
		product, err := fetcher.Fetch(flag.Arg(0))
		if err != nil {
			logger.Fatalln(err)
		}
		pkg := recycleme.NewProductPackage(product)
		if *jsonFlag {
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
