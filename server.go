package recycleme

import (
	"encoding/json"
	"fmt"
	eancheck "github.com/nicholassm/go-ean"
	"log"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/index.html")
}

func MaterialsHandler(w http.ResponseWriter, r *http.Request, db MaterialDB) {
	materials, err := db.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(materials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", out)
}

func AddBlacklistHandler(b BlacklistDB, w http.ResponseWriter, r *http.Request, logger *log.Logger, f Fetcher, m Mailer) {
	r.ParseForm()
	url := r.FormValue("url")
	ean := r.FormValue("ean")

	if url == "" || ean == "" {
		http.Error(w, "missing form data", http.StatusInternalServerError)
		return
	}

	if !f.IsURLValidForEAN(url, ean) {
		msg := fmt.Sprintf("url %v invalid for ean %v", url, ean)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	b.Add(url)
	name := r.FormValue("name")
	logger.Println(fmt.Sprintf("Blacklisting %s. %s should be %s", url, ean, name))
	fmt.Fprintf(w, "added")
	go func() {
		err := m(ean+" blacklisted", fmt.Sprintf("Blacklisting %s.\n%s should be %s", url, ean, name))
		if err != nil {
			logger.Println(err)
		}
	}()
}

func AddPackageHandler(db PackagesDB, w http.ResponseWriter, r *http.Request, logger *log.Logger, m Mailer) {
	r.ParseForm()
	materialsStr := r.FormValue("materials")
	ean := r.FormValue("ean")
	if materialsStr == "" || ean == "" {
		http.Error(w, "missing form data", http.StatusInternalServerError)
		return
	}

	if !eancheck.Valid(ean) {
		msg := fmt.Sprintf("invalid ean %v", ean)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var materials []Material
	err := json.Unmarshal([]byte(materialsStr), &materials)
	if err != nil {
		msg := fmt.Sprintf("invalid materials format %v for %v", materials, ean)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	db.Set(ean, materials)
	logger.Println(fmt.Sprintf("Adding %v for %v", materials, ean))
	fmt.Fprintf(w, "added")
	go func() {
		err := m("Adding package for "+ean, fmt.Sprintf("Materials added to %v:\n%v", ean, materials))
		if err != nil {
			logger.Println(err)
		}
	}()
}

func ThrowAwayHandler(db PackagesDB, blacklistDB BlacklistDB, w http.ResponseWriter, r *http.Request, f Fetcher) {
	ean := r.URL.Path[len("/throwaway/"):]
	product, err := f.Fetch(ean, blacklistDB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg, err := NewProductPackage(product, db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonBytes, err := pkg.ThrowAwayJSON(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", jsonBytes)
}
