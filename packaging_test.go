package recycleme

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

var binsJSON = `{
  "Bins": [
    {
      "id": 0,
      "name": "Bac à couvercle vert"
    },
    {
      "id": 1,
      "name": "Bac à couvercle jaune"
    },
    {
      "id": 2,
      "name": "Bac à couvercle blanc"
    }
  ]
}`

var materialsJSON = `{
  "Materials": [
    {
      "id": 0,
      "name": "Boîte carton",
      "binId": 1
    },
    {
      "id": 1,
      "name": "Film plastique",
      "binId": 0
    },
    {
      "id": 2,
      "name": "Bouteille plastique",
      "binId": 1
    },
    {
      "id": 3,
      "name": "Bouteille de verre",
      "binId": 2
    },
    {
      "id": 4,
      "name": "Nourriture",
      "binId": 0
    },
    {
      "id": 5,
      "name": "Bouchon de bouteille en plastique",
      "binId": 1
    },
    {
      "id": 6,
      "name": "Bouchon de bouteille en métal",
      "binId": 1
    },
    {
      "id": 7,
      "name": "Kleenex",
      "binId": 0
    },
    {
      "id": 8,
      "name": "Boîte plastique",
      "binId": 0
    }
  ]
}`

var packagesJSON = `{
  "Packages": [
    {
      "id": 0,
      "ean": "7613034383808",
      "materialIds": [
        0,
        1,
        4
      ]
    },
    {
      "id": 1,
      "ean": "5029053038896",
      "materialIds": [
        0,
        7
      ]
    },
    {
      "id": 2,
      "ean": "3281780874976",
      "materialIds": [
        1,
        8
      ]
    }
  ]
}`

func init() {
	logger := log.New(ioutil.Discard, "", 0)
	LoadBinsJSON(strings.NewReader(binsJSON), logger)
	LoadMaterialsJSON(strings.NewReader(materialsJSON), logger)
	LoadPackagesJSON(strings.NewReader(packagesJSON), logger)
}

func TestPackage(t *testing.T) {
	r := Package{EAN: "7613034383808", Materials: []Material{
		Material{ID: 0, Name: "Boîte carton"},
		Material{ID: 1, Name: "Film plastique"},
		Material{ID: 4, Name: "Nourriture"},
	}}

	pkg, err := Packages.Get(r.EAN)
	if err != nil {
		if err == ErrPackageNotFound {
			t.Fatalf("No package found for %v", r.EAN)
		} else {
			t.Fatal(err)
		}
	}
	if r.EAN != pkg.EAN || len(r.Materials) != len(pkg.Materials) {
		t.Fatalf("Packages for %v differ", r.EAN)
	}

	for i, m := range r.Materials {
		pkgMaterial := pkg.Materials[i]
		if m.ID != pkgMaterial.ID || m.Name != pkgMaterial.Name {
			t.Errorf("Material differ for EAN %v: %v vs %v", r.EAN, m, pkgMaterial)
		}
	}

	binNames := []string{"Bac à couvercle jaune", "Bac à couvercle vert", "Bac à couvercle vert"}
	for i, m := range r.Materials {
		bin := MaterialsToBins[m]
		if bin.Name != binNames[i] {
			t.Errorf("material %v belong to %v, not %v", m.Name, binNames[i], bin.Name)
		}
	}
}

func TestProductPackage(t *testing.T) {
	product := Product{EAN: "7613034383808", Name: "Four à Pierre Royale", URL: "http://fr.openfoodfacts.org/api/v0/produit/7613034383808.json", ImageURL: "http://static.openfoodfacts.org/images/products/761/303/438/3808/front.8.400.jpg"}
	pp, err := NewProductPackage(product, Packages)
	if err != nil {
		t.Fatal(err)
	}
	materials := []Material{
		Material{ID: 0, Name: "Boîte carton"},
		Material{ID: 1, Name: "Film plastique"},
		Material{ID: 4, Name: "Nourriture"}}
	if pp.Product != product {
		t.Errorf("Some attributes are invalid for: %v; expected %v", pp, product)
	}

	if len(pp.Materials) != len(materials) {
		t.Fatalf("Packages for %v differ", pp.EAN)
	}
	for i, m := range materials {
		pkgMaterial := pp.Materials[i]
		if m.ID != pkgMaterial.ID || m.Name != pkgMaterial.Name {
			t.Errorf("Material differ for EAN %v: %v vs %v", pp.EAN, m, pkgMaterial)
		}
	}

	binNames := map[string][]string{"Bac à couvercle jaune": []string{"Boîte carton"}, "Bac à couvercle vert": []string{"Film plastique", "Nourriture"}}
	for bin, ms := range pp.ThrowAway() {
		for _, m := range ms {
			found := false
			for _, m2 := range binNames[bin.Name] {
				if m.Name == m2 {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("material %v not found in %v", m.Name, bin.Name)
			}
		}
	}
}

func TestThrowAwayJSON(t *testing.T) {
	product := Product{EAN: "7613034383808"}
	pkg, err := NewProductPackage(product, Packages)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal(throwAwaypackage{
		Product: ProductPackage{
			Product:   product,
			Materials: []Material{Materials[0], Materials[1], Materials[4]}},
		ThrowAway: map[string][]Material{Bins[1].Name: []Material{Materials[0]}, Bins[0].Name: []Material{Materials[1], Materials[4]}},
	})
	if err != nil {
		t.Fatal(err)
	}

	out, err := pkg.ThrowAwayJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(expected) {
		t.Errorf("ThrowAwayJson not as expected: %v != %v", string(out), string(expected))
	}
}
