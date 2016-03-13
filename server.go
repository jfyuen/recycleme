package recycleme

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func HomeHandler(_ *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("data/index.html")
		//err := templates.ExecuteTemplate(w, "index.html", struct{}{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		t.Execute(w, struct{}{})
	}
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

func PackageHandler(w http.ResponseWriter, r *http.Request) {
	ean := r.URL.Path[len("/package/"):]
	product, err := Scrap(ean)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg := NewProductPackage(product)
	b, err := json.Marshal(pkg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", b)
}

func ThrowAwayHandler(w http.ResponseWriter, r *http.Request) {
	ean := r.URL.Path[len("/throwaway/"):]
	product, err := Scrap(ean)
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
