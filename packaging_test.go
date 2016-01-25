package recycleme

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

var binsJson = `{
  "Bins": [
    {
      "id": 0,
      "Name": "Green Bin"
    },
    {
      "id": 1,
      "Name": "Yellow Bin"
    },
    {
      "id": 2,
      "Name": "White Bin"
    }
  ]
}`

var materialsJson = `{
  "Materials": [
    {
      "id": 0,
      "Name": "Cardboard box",
      "binIds": [1]
    },
    {
      "id": 1,
      "Name": "Plastic foil",
      "binIds": [0]
    },
    {
      "id": 2,
      "Name": "Plastic bottle",
      "binIds": [1]
    },
    {
      "id": 3,
      "Name": "Glass bottle",
      "binIds": [2]
    },
    {
      "id": 4,
      "Name": "Food",
      "binIds": [0]
    },
    {
      "id": 5,
      "Name": "Plastic bottle cap",
      "binIds": [1]
    },
    {
      "id": 6,
      "Name": "Metal bottle cap",
      "binIds": [1]
    },
    {
      "id": 7,
      "Name": "Kleenex",
      "binIds": [0]
    },
    {
      "id": 8,
      "Name": "Plastic box",
      "binIds": [0]
    }
  ]
}`

var packagesJson = `{
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
	LoadBinsJson(strings.NewReader(binsJson), logger)
	LoadMaterialsJson(strings.NewReader(materialsJson), logger)
	LoadPackagesJson(strings.NewReader(packagesJson), logger)
}

func TestPackage(t *testing.T) {
	r := Package{EAN: "7613034383808", Materials: []Material{
		Material{id: 0, Name: "Cardboard box"},
		Material{id: 1, Name: "Plastic foil"},
		Material{id: 4, Name: "Food"},
	}}

	pkg := Packages["7613034383808"]
	if r.EAN != pkg.EAN || len(r.Materials) != len(pkg.Materials) {
		t.Errorf("Packages for %v differ", r.EAN)
	} else {
		for i, m := range r.Materials {
			pkgMaterial := pkg.Materials[i]
			if m.id != pkgMaterial.id || m.Name != pkgMaterial.Name {
				t.Errorf("Material differ for EAN %v: %v vs %v", r.EAN, m, pkgMaterial)
			}
		}

		binNames := []string{"Yellow Bin", "Green Bin", "Green Bin"}
		for i, m := range r.Materials {
			bins := MaterialsToBins[m]
			if len(bins) != 1 {
				t.Errorf("material %v must belongs to 1 bin, found %v", m.Name, len(bins))
			} else {
				if bins[0].Name != binNames[i] {
					t.Errorf("material %v belong to %v, not %v", m.Name, binNames[i], bins[0].Name)
				}
			}
		}
	}
}

func TestProductPackage(t *testing.T) {
	product := Product{EAN: "7613034383808", Name: "Four à Pierre Royale", URL: "http://fr.openfoodfacts.org/api/v0/produit/7613034383808.json", ImageURL: "http://static.openfoodfacts.org/images/products/761/303/438/3808/front.8.400.jpg"}
	pp := NewProductPackage(product)
	materials := []Material{
		Material{id: 0, Name: "Cardboard box"},
		Material{id: 1, Name: "Plastic foil"},
		Material{id: 4, Name: "Food"}}
	if pp.Product != product {
		t.Errorf("Some attributes are invalid for: %v; expected %v", pp, product)
	}

	if len(pp.materials) != len(materials) {
		t.Errorf("Packages for %v differ", pp.EAN)
	} else {
		for i, m := range materials {
			pkgMaterial := pp.materials[i]
			if m.id != pkgMaterial.id || m.Name != pkgMaterial.Name {
				t.Errorf("Material differ for EAN %v: %v vs %v", pp.EAN, m, pkgMaterial)
			}
		}

		binNames := map[string]string{"Cardboard box": "Yellow Bin", "Plastic foil": "Green Bin", "Food": "Green Bin"}
		i := 0
		for m, bins := range pp.ThrowAway() {
			if len(bins) != 1 {
				t.Errorf("material %v must belongs to 1 bin, found %v", m.Name, len(bins))
			} else {
				if bins[0].Name != binNames[m.Name] {
					t.Errorf("material %v belong to %v, not %v", m.Name, binNames[m.Name], bins[0].Name)
				}
			}
			i++
		}
	}
}

func TestThrowAwayJson(t *testing.T) {
	product := Product{EAN: "7613034383808"}
	pkg := NewProductPackage(product)
	expected := `{"Cardboard box":[{"Name":"Yellow Bin"}],"Food":[{"Name":"Green Bin"}],"Plastic foil":[{"Name":"Green Bin"}]}`
	out, err := pkg.ThrowAwayJson()
	if err != nil {
		t.Error(err)
	} else {
		if string(out) != expected {
			t.Errorf("ThrowAwayJson not as expected: %v %v", string(out), expected)
		}
	}
}