package recycleme

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/index.html")
}

func BinHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/bin/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bin, ok := Bins[id]
	if !ok {
		http.Error(w, fmt.Sprintf("bin id %d not found", id), http.StatusNotFound)
		return
	}
	out, err := json.Marshal(bin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", out)
}

func BinsHandler(w http.ResponseWriter, r *http.Request) {
	tmpMap := make(map[string]Bin)
	for id, b := range Bins {
		idStr := strconv.Itoa(id)
		tmpMap[idStr] = b
	}
	out, err := json.Marshal(tmpMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", out)
}

func MaterialsHandler(w http.ResponseWriter, r *http.Request) {
	tmpMap := make(map[string]Material)
	for id, m := range Materials {
		idStr := strconv.Itoa(id)
		tmpMap[idStr] = m
	}
	out, err := json.Marshal(tmpMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", out)
}

func (b *blacklist) AddBlacklistHandler(w http.ResponseWriter, r *http.Request, logger *log.Logger, f Fetcher) {
	r.ParseForm()
	url := r.FormValue("url")
	ean := r.FormValue("ean")

	if f.IsURLValidForEAN(url, ean) {
		b.Add(url)
		name := r.FormValue("name")
		logger.Printf("Blacklisting %s. %s should be %s\n", url, ean, name)
	}
	fmt.Fprintf(w, "ok")
}

func ThrowAwayHandler(w http.ResponseWriter, r *http.Request, f Fetcher) {
	ean := r.URL.Path[len("/throwaway/"):]
	product, err := f.Fetch(ean)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg := NewProductPackage(product)
	jsonBytes, err := pkg.ThrowAwayJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", jsonBytes)
}
