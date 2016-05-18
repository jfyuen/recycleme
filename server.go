package recycleme

import (
	"encoding/json"
	"fmt"
	eancheck "github.com/nicholassm/go-ean"
	"log"
	"net/http"
)

type HomeHandler struct{}

func (h HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "static/index.html")
	} else {
		http.Error(w, "page not found", http.StatusNotFound)
	}
}

type MaterialsHandler struct {
	DB MaterialDB
}

func (m MaterialsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	materials, err := m.DB.GetAll()
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

type AddBlacklistHandler struct {
	Logger    *log.Logger
	Fetcher   Fetcher
	Mailer    Mailer
	Blacklist BlacklistDB
}

func (h AddBlacklistHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.FormValue("url")
	ean := r.FormValue("ean")

	if url == "" || ean == "" {
		http.Error(w, "missing form data", http.StatusInternalServerError)
		return
	}

	if !h.Fetcher.IsURLValidForEAN(url, ean) {
		msg := fmt.Sprintf("url %v invalid for ean %v", url, ean)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	h.Blacklist.Add(url)
	name := r.FormValue("name")
	h.Logger.Println(fmt.Sprintf("Blacklisting %s. %s should be %s", url, ean, name))
	fmt.Fprintf(w, "added")
	go func() {
		err := h.Mailer(ean+" blacklisted", fmt.Sprintf("Blacklisting %s.\n%s should be %s", url, ean, name))
		if err != nil {
			h.Logger.Println(err)
		}
	}()
}

type AddPackageHandler struct {
	DB     PackagesDB
	Logger *log.Logger
	Mailer Mailer
}

func (h AddPackageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	h.DB.Set(ean, materials)
	h.Logger.Println(fmt.Sprintf("Adding %v for %v", materials, ean))
	fmt.Fprintf(w, "added")
	go func() {
		err := h.Mailer("Adding package for "+ean, fmt.Sprintf("Materials added to %v:\n%v", ean, materials))
		if err != nil {
			h.Logger.Println(err)
		}
	}()
}

type ThrowAwayHandler struct {
	DB          PackagesDB
	BlacklistDB BlacklistDB
	Fetcher     Fetcher
}

func (h ThrowAwayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ean := r.URL.Path[len("/throwaway/"):]
	product, err := h.Fetcher.Fetch(ean, h.BlacklistDB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg, err := NewProductPackage(product, h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonBytes, err := pkg.ThrowAwayJSON(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", jsonBytes)
}
