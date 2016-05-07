package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/jfyuen/recycleme"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

var jsonFlag = flag.Bool("json", false, "Print json export")
var serverFlag = flag.Bool("server", false, "Run in server mode, serving json (EAN as input is useless)")
var serverPort = flag.String("p", "8080", "Port to listen to")

func init() {
	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s -d DIR [options] EAN:\n", name)
		flag.PrintDefaults()
	}
}

func NewMgoDB(url string) (*mgo.Session, error) {

	if url == "" {
		return nil, errors.New("invalid mongodb connection parameters")
	}
	timeout := 60 * time.Second
	mongoSession, err := mgo.DialWithTimeout(url, timeout)
	return mongoSession, err
}

func main() {
	flag.Parse()
	if (len(flag.Args()) != 1 && !*serverFlag) || (*serverFlag && len(flag.Args()) != 0) {
		flag.Usage()
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	mongoSession, err := NewMgoDB(os.Getenv("RECYCLEME_MONGO_URI"))
	if err != nil {
		logger.Fatal(err)
	}
	defer mongoSession.Close()

	packageDB := recycleme.NewMgoPackageDB(mongoSession, "")
	blacklistDB := recycleme.NewMgoBlacklistDB(mongoSession, "")

	localProductDB := recycleme.NewMgoLocalProductDB(mongoSession, "")
	fetcher, err := recycleme.NewDefaultFetcher(localProductDB)
	if err != nil {
		logger.Println(err.Error())
	}

	if *serverFlag {
		emailConfig, err := recycleme.NewEmailConfig(os.Getenv("RECYCLEME_MAIL_HOST"), os.Getenv("RECYCLEME_MAIL_RECIPIENT"), os.Getenv("RECYCLEME_MAIL_USERNAME"), os.Getenv("RECYCLEME_MAIL_PASSWORD"))
		var mailHandler recycleme.Mailer
		if err != nil {
			logger.Println(err.Error())
			mailHandler = func(subject, body string) error {
				return nil
			}
		} else {
			mailHandler = emailConfig.SendMail
		}
		http.HandleFunc("/materials/", func(w http.ResponseWriter, r *http.Request) {
			recycleme.MaterialsHandler(w, r, packageDB)
		})
		http.HandleFunc("/package/add", func(w http.ResponseWriter, r *http.Request) {
			recycleme.AddPackageHandler(packageDB, w, r, logger, mailHandler)
		})
		http.HandleFunc("/blacklist/add", func(w http.ResponseWriter, r *http.Request) {
			recycleme.AddBlacklistHandler(blacklistDB, w, r, logger, fetcher, mailHandler)
		})
		http.HandleFunc("/throwaway/", func(w http.ResponseWriter, r *http.Request) {
			recycleme.ThrowAwayHandler(packageDB, blacklistDB, w, r, fetcher)
		})
		http.HandleFunc("/", recycleme.HomeHandler)

		fs := http.FileServer(http.Dir("static"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))

		logger.Println("Running in server mode on port " + *serverPort)
		err = http.ListenAndServe(":"+*serverPort, nil)
		if err != nil {
			logger.Fatalln(err)
		}
	} else {
		product, err := fetcher.Fetch(flag.Arg(0), blacklistDB)
		if err != nil {
			logger.Fatalln(err)
		}
		pkg, err := recycleme.NewProductPackage(product, packageDB)
		if err != nil {
			logger.Fatalln(err)
		}
		throwaway, err := pkg.ThrowAway(packageDB)
		if err != nil {
			logger.Fatalln(err)
		}
		if *jsonFlag {
			jsonBytes, err := json.Marshal(throwaway)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Println(string(jsonBytes))
		} else {
			logger.Println(throwaway)
		}
	}
}
