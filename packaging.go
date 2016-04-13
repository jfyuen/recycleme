package recycleme

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

var ErrPackageNotFound = errors.New("ean not found in packages db")

type Bin struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// Bins: id -> Bin
var Bins = make(map[int]Bin)

// MaterialsToBins: Material -> Bin, only for Paris (France) at the moment
var MaterialsToBins = make(map[Material]Bin)

// Materials: id -> Material
var Materials = make(map[int]Material)

// A Material composes Packaging, different Materials go to different Bin, event ones that may be close enough
// For example, in Paris, plastic bags go to the green bin, but plastic bottles go to the yellow bin
type Material struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Packaging related to a Product
// Several products may have the same types of packaging
// For example, a pizza box and a frozen product may both have a cardboard box and a plastic foil
type Package struct {
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

type memoryPackagesDB struct {
	byEAN map[string]Package
	sync.RWMutex
}

type PackagesDB interface {
	Get(ean string) (Package, error)
	Set(ean string, m []Material) error
}

func (p *memoryPackagesDB) Set(ean string, m []Material) error {
	pack := Package{EAN: ean, Materials: m}
	p.Lock()
	p.byEAN[ean] = pack
	p.Unlock()
	return nil
}

func (p *memoryPackagesDB) Get(ean string) (Package, error) {
	p.RLock()
	defer p.RUnlock()
	v, ok := p.byEAN[ean]
	if !ok {
		return Package{}, ErrPackageNotFound
	}
	return v, nil
}

func newPackages() *memoryPackagesDB {
	return &memoryPackagesDB{byEAN: make(map[string]Package)}
}

// Packages indexes Package by EAN
var Packages = newPackages()

// ProductPackage links a Product and its packages
type ProductPackage struct {
	Product
	Materials []Material `json:"materials"`
}

func NewProductPackage(p Product, db PackagesDB) (ProductPackage, error) {
	pp := ProductPackage{Product: p}
	if pkg, err := db.Get(p.EAN); err != nil {
		if err == ErrPackageNotFound {
			pp.Materials = make([]Material, 0, 0)
			return pp, nil
		}
		return pp, err
	} else {
		pp.Materials = pkg.Materials
	}
	return pp, nil
}

func (pp ProductPackage) ThrowAway() map[Bin][]Material {
	bins := make(map[Bin][]Material)
	for _, m := range pp.Materials {
		bin := MaterialsToBins[m]
		lst := bins[bin]
		bins[bin] = append(lst, m)
	}
	return bins
}

type throwAwaypackage struct {
	Product   ProductPackage        `json:"product"`
	ThrowAway map[string][]Material `json:"throwAway"`
}

func (pp ProductPackage) ThrowAwayJSON() ([]byte, error) {
	throwAway := pp.ThrowAway()
	out := make(map[string][]Material)
	for k, v := range throwAway {
		out[k.Name] = v
	}
	return json.Marshal(throwAwaypackage{pp, out})
}

func readJSON(r io.Reader, logger *log.Logger) interface{} {
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

func LoadBinsJSON(r io.Reader, logger *log.Logger) {
	jsonBins := readJSON(r, logger)
	bins := jsonBins.(map[string]interface{})
	for _, mIntf := range bins["Bins"].([]interface{}) {
		m := mIntf.(map[string]interface{})
		id := m["id"].(float64)
		bin := Bin{ID: int(id), Name: m["name"].(string)}
		Bins[bin.ID] = bin
	}
}

func LoadMaterialsJSON(r io.Reader, logger *log.Logger) {
	jsonMaterials := readJSON(r, logger)
	materials := jsonMaterials.(map[string]interface{})
	for _, mIntf := range materials["Materials"].([]interface{}) {
		m := mIntf.(map[string]interface{})
		id := m["id"].(float64)
		material := Material{ID: int(id), Name: m["name"].(string)}
		binID := m["binId"].(float64)
		bin, ok := Bins[int(binID)]
		if !ok {
			logger.Fatal(fmt.Errorf("binId %v not found in Bins %v", binID, Bins))
		}
		MaterialsToBins[material] = bin
		Materials[material.ID] = material
	}
}

func LoadPackagesJSON(r io.Reader, logger *log.Logger) {
	jsonMaterials := readJSON(r, logger)
	materials := jsonMaterials.(map[string]interface{})
	for _, mIntf := range materials["Packages"].([]interface{}) {
		m := mIntf.(map[string]interface{})
		materialsIds := m["materialIds"].([]interface{})
		var materials []Material
		for i := range materialsIds {
			materialID := int(materialsIds[i].(float64))
			material, ok := Materials[materialID]
			if !ok {
				logger.Fatal(fmt.Errorf("materialId %v not found in Materials %v", materialID, Materials))
			}
			materials = append(materials, material)
		}
		ean := m["ean"].(string)
		Packages.Set(ean, materials)
	}
}

func LoadBlacklistJSON(r io.Reader, logger *log.Logger) {
	jsonBlacklist := readJSON(r, logger)
	blacklist := jsonBlacklist.(map[string]interface{})
	for _, url := range blacklist["Blacklist"].([]interface{}) {
		Blacklist.Add(url.(string))
	}

}

func LoadJSONFiles(dir string, logger *log.Logger) {
	files := []string{"bins.json", "materials.json", "packages.json", "blacklist.json"}
	funcs := []func(io.Reader, *log.Logger){LoadBinsJSON, LoadMaterialsJSON, LoadPackagesJSON, LoadBlacklistJSON}
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
