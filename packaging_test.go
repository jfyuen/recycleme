package recycleme

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

var binsJSON = `{
  "Bins": [
    {
      "id": 0,
      "Name": "Bac à couvercle vert"
    },
    {
      "id": 1,
      "Name": "Bac à couvercle jaune"
    },
    {
      "id": 2,
      "Name": "Bac à couvercle blanc"
    }
  ]
}`

var materialsJSON = `{
  "Materials": [
    {
      "id": 0,
      "Name": "Boîte carton",
      "binId": 1
    },
    {
      "id": 1,
      "Name": "Film plastique",
      "binId": 0
    },
    {
      "id": 2,
      "Name": "Bouteille plastique",
      "binId": 1
    },
    {
      "id": 3,
      "Name": "Bouteille de verre",
      "binId": 2
    },
    {
      "id": 4,
      "Name": "Nourriture",
      "binId": 0
    },
    {
      "id": 5,
      "Name": "Bouchon de bouteille en plastique",
      "binId": 1
    },
    {
      "id": 6,
      "Name": "Bouchon de bouteille en métal",
      "binId": 1
    },
    {
      "id": 7,
      "Name": "Kleenex",
      "binId": 0
    },
    {
      "id": 8,
      "Name": "Boîte plastique",
      "binId": 0
    }
  ]
}`

var packagesJSON = `{
  "Packages": [
    {
      "id": 0,
      "EAN": "7613034383808",
      "materialIds": [
        0,
        1,
        4
      ]
    },
    {
      "id": 1,
      "EAN": "5029053038896",
      "materialIds": [
        0,
        7
      ]
    },
    {
      "id": 2,
      "EAN": "3281780874976",
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
		Material{Id: 0, Name: "Boîte carton"},
		Material{Id: 1, Name: "Film plastique"},
		Material{Id: 4, Name: "Nourriture"},
	}}

	pkg, ok := Packages.Get(r.EAN)
	if !ok {
		t.Fatalf("No package found for %v", r.EAN)
	}
	if r.EAN != pkg.EAN || len(r.Materials) != len(pkg.Materials) {
		t.Fatalf("Packages for %v differ", r.EAN)
	}

	for i, m := range r.Materials {
		pkgMaterial := pkg.Materials[i]
		if m.Id != pkgMaterial.Id || m.Name != pkgMaterial.Name {
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
	pp, err := NewProductPackage(product)
	if err != nil {
		t.Fatal(err)
	}
	materials := []Material{
		Material{Id: 0, Name: "Boîte carton"},
		Material{Id: 1, Name: "Film plastique"},
		Material{Id: 4, Name: "Nourriture"}}
	if pp.Product != product {
		t.Errorf("Some attributes are invalid for: %v; expected %v", pp, product)
	}

	if len(pp.materials) != len(materials) {
		t.Fatalf("Packages for %v differ", pp.EAN)
	}
	for i, m := range materials {
		pkgMaterial := pp.materials[i]
		if m.Id != pkgMaterial.Id || m.Name != pkgMaterial.Name {
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
	pkg, err := NewProductPackage(product)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"Product":{"EAN":"7613034383808","Name":"","URL":"","ImageURL":"","WebsiteURL":"","WebsiteName":""},"ThrowAway":{"Bac à couvercle jaune":[{"Id":0,"Name":"Boîte carton"}],"Bac à couvercle vert":[{"Id":1,"Name":"Film plastique"},{"Id":4,"Name":"Nourriture"}]}}`
	out, err := pkg.ThrowAwayJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != expected {
		t.Errorf("ThrowAwayJson not as expected: %v != %v", string(out), expected)
	}
}
