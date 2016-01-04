package recycleme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type Bin struct {
	Name string
	id   int
}

// Bins: id -> Bin
var Bins = make(map[int]Bin)

// MaterialsToBins: Material -> []Bin, in all countries (to be filtered later)
var MaterialsToBins = make(map[Material][]Bin)

// Materials: id -> Material
var Materials = make(map[int]Material)

// A Material composes Packaging, different Materials go to different Bin, event ones that may be close enough
// For example, in Paris, plastic bags go to the green bin, but plastic bottles go to the yellow bin
type Material struct {
	id   int
	Name string
}

// Packaging related to a Product
// Several products may have the same types of packaging
// For example, a pizza box and a frozen product may both have a cardboard box and a plastic foil
type Package struct {
	id        int
	EAN       string
	Materials []Material
}

func (p Package) String() string {
	s := make([]string, len(p.Materials), len(p.Materials))
	for i, m := range p.Materials {
		s[i] = m.Name
	}
	return fmt.Sprintf("Product %v is composed of %v", p.EAN, strings.Join(s, ", "))
}

// Packages indexes Package by EAN
var Packages = make(map[string]Package)

// ProductPackage Link a Product and its packages
type ProductPackage struct {
	*Product
	materials []Material
}

func NewProductPackage(p *Product) ProductPackage {
	return ProductPackage{Product: p}
}

func (pp ProductPackage) Bins() []Material {
	return nil
}

func readJson(r io.Reader, logger *log.Logger) interface{} {
	var data interface{}
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if n == 0 {
		logger.Fatal(fmt.Errorf("nothing was read from Reader"))
	}

	if err != nil {
		logger.Fatal(err)
	}
	err = json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		logger.Fatal(err)
	}
	return data
}

func LoadBinsJson(r io.Reader, logger *log.Logger) {
	jsonBins := readJson(r, logger)
	bins := jsonBins.(map[string]interface{})
	for _, mIntf := range bins["Bins"].([]interface{}) {
		m := mIntf.(map[string]interface{})
		id := m["id"].(float64)
		bin := Bin{id: int(id), Name: m["Name"].(string)}
		Bins[bin.id] = bin
	}
}

func LoadMaterialsJson(r io.Reader, logger *log.Logger) {
	jsonMaterials := readJson(r, logger)
	materials := jsonMaterials.(map[string]interface{})
	for _, mIntf := range materials["Materials"].([]interface{}) {
		m := mIntf.(map[string]interface{})
		id := m["id"].(float64)
		material := Material{id: int(id), Name: m["Name"].(string)}
		bins := MaterialsToBins[material]
		binIds := m["binIds"].([]interface{})
		for i := range binIds {
			binId := int(binIds[i].(float64))
			bin, ok := Bins[binId]
			if !ok {
				logger.Fatal(fmt.Errorf("binId %v not found in Bins %v", binId, Bins))
			}
			bins = append(bins, bin)
		}
		Materials[material.id] = material
	}
}

func LoadPackagesJson(r io.Reader, logger *log.Logger) {
	jsonMaterials := readJson(r, logger)
	materials := jsonMaterials.(map[string]interface{})
	for _, mIntf := range materials["Packages"].([]interface{}) {
		m := mIntf.(map[string]interface{})
		id := int(m["id"].(float64))
		pkg := Package{id: int(id), EAN: m["EAN"].(string)}
		materialsIds := m["materialIds"].([]interface{})
		for i := range materialsIds {
			materialId := int(materialsIds[i].(float64))
			material, ok := Materials[materialId]
			if !ok {
				logger.Fatal(fmt.Errorf("materialId %v not found in Materials %v", materialId, Materials))
			}
			pkg.Materials = append(pkg.Materials, material)
		}
		Packages[pkg.EAN] = pkg
	}
}

func LoadJsonFiles(dir string, logger *log.Logger) {
	files := []string{"bins.json", "materials.json", "packages.json"}
	funcs := []func(io.Reader, *log.Logger){LoadBinsJson, LoadMaterialsJson, LoadPackagesJson}
	for i, filename := range files {
		path := path.Join(dir, filename)
		f, err := os.Open(path)
		if err != nil {
			logger.Fatal(err)
		}
		defer f.Close()
		funcs[i](f, logger)
	}
}
