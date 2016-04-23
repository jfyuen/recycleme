package recycleme

import (
	"encoding/json"
	"errors"
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"testing"
	"time"
)

var binsJSON = `[
    {
      "id": 1,
      "name": "Bac à couvercle vert"
    },
    {
      "id": 2,
      "name": "Bac à couvercle jaune"
    },
    {
      "id": 3,
      "name": "Bac à couvercle blanc"
    }
]`

var materialsJSON = `[
    {
      "id": 1,
      "name": "Boîte carton",
      "bin_id": 2
    },
    {
      "id": 2,
      "name": "Film plastique",
      "bin_id": 1
    },
    {
      "id": 3,
      "name": "Bouteille plastique",
      "bin_id": 2
    },
    {
      "id": 4,
      "name": "Bouteille de verre",
      "bin_id": 3
    },
    {
      "id": 5,
      "name": "Nourriture",
      "bin_id": 1
    },
    {
      "id": 6,
      "name": "Bouchon de bouteille en plastique",
      "bin_id": 2
    },
    {
      "id": 7,
      "name": "Bouchon de bouteille en métal",
      "bin_id": 2
    },
    {
      "id": 8,
      "name": "Kleenex",
      "bin_id": 1
    },
    {
      "id": 9,
      "name": "Boîte plastique",
      "bin_id": 1
    }
]`

var packagesJSON = `[
    {
      "ean": "7613034383808",
      "material_ids": [
        1,
        2,
        5
      ]
    },
    {
      "ean": "5029053038896",
      "material_ids": [
        1,
        8
      ]
    },
    {
      "ean": "3281780874976",
      "material_ids": [
        2,
        9
      ]
    }
]`

func NewMgoDB(url string) (*mgo.Session, error) {

	if url == "" {
		return nil, errors.New("invalid mongodb connection parameters")
	}
	timeout := 60 * time.Second
	mongoSession, err := mgo.DialWithTimeout(url, timeout)
	return mongoSession, err
}

func createBins(db *mgoPackagesDB) error {
	return withMgoSession(db.session, func(s *mgo.Session) error {
		var bins []Bin
		if err := json.Unmarshal([]byte(binsJSON), &bins); err != nil {
			return err
		}

		collection := s.DB("").C(db.binsColName)
		if err := collection.DropCollection(); err != nil && err.Error() != "ns not found" {
			return err
		}

		for _, bin := range bins {
			if err := collection.Insert(bin); err != nil {
				return err
			}
		}

		return nil
	})
}

func createMaterials(db *mgoPackagesDB) error {
	return withMgoSession(db.session, func(s *mgo.Session) error {
		var materialsWithBinId []struct {
			Material `json:",inline"`
			BinID    uint `json:"bin_id"`
		}
		if err := json.Unmarshal([]byte(materialsJSON), &materialsWithBinId); err != nil {
			return err
		}

		materialsCols := s.DB("").C(db.materialsColName)
		if err := materialsCols.DropCollection(); err != nil && err.Error() != "ns not found" {
			return err
		}

		materialsToBinsCols := s.DB("").C(db.materialsToBinsColName)
		if err := materialsToBinsCols.DropCollection(); err != nil && err.Error() != "ns not found" {
			return err
		}

		for _, m := range materialsWithBinId {
			if err := materialsCols.Insert(m.Material); err != nil {
				return err
			}
			v := struct {
				MaterialdID uint `bson:"material_id"`
				BinID       uint `bson:"bin_id"`
			}{m.ID, m.BinID}
			if err := materialsToBinsCols.Insert(v); err != nil {
				return err
			}
		}

		return nil
	})
}

func createPackages(db *mgoPackagesDB) error {
	return withMgoSession(db.session, func(s *mgo.Session) error {
		var packages []mgoPackageItem
		if err := json.Unmarshal([]byte(packagesJSON), &packages); err != nil {
			return err
		}

		collection := s.DB("").C(db.packagesColName)
		if err := collection.DropCollection(); err != nil && err.Error() != "ns not found" {
			return err
		}

		for _, p := range packages {
			if err := collection.Insert(p); err != nil {
				return err
			}
		}

		return nil
	})
}

func TestPackage(t *testing.T) {
	r := Package{EAN: "7613034383808", Materials: []Material{
		Material{ID: 1, Name: "Boîte carton"},
		Material{ID: 2, Name: "Film plastique"},
		Material{ID: 5, Name: "Nourriture"},
	}}

	pkg, err := packageDB.Get(r.EAN)
	if err != nil {
		if err == errPackageNotFound {
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

	materialsToBins, err := packageDB.GetBins(r.Materials)
	if err != nil {
		t.Fatal(err)
	}

	binNames := []string{"Bac à couvercle jaune", "Bac à couvercle vert", "Bac à couvercle vert"}
	for i, m := range r.Materials {
		bin := materialsToBins[m]
		if bin.Name != binNames[i] {
			t.Errorf("material %v belong to %v, not %v", m.Name, binNames[i], bin.Name)
		}
	}
}

func TestProductPackage(t *testing.T) {
	product := Product{EAN: "7613034383808", Name: "Four à Pierre Royale", URL: "http://fr.openfoodfacts.org/api/v0/produit/7613034383808.json", ImageURL: "http://static.openfoodfacts.org/images/products/761/303/438/3808/front.8.400.jpg"}
	pp, err := NewProductPackage(product, packageDB)
	if err != nil {
		t.Fatal(err)
	}
	materials := []Material{
		Material{ID: 1, Name: "Boîte carton"},
		Material{ID: 2, Name: "Film plastique"},
		Material{ID: 5, Name: "Nourriture"}}
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

	binNames := map[string]string{"Boîte carton": "Bac à couvercle jaune", "Film plastique": "Bac à couvercle vert", "Nourriture": "Bac à couvercle vert"}
	materialToBins, err := pp.ThrowAway(packageDB)
	if err != nil {
		t.Fatal(err)
	}
	for material, bin := range materialToBins {
		expected := binNames[material.Name]
		if bin.Name != expected {
			t.Errorf("invalid bin (%v) for material %v, expected %v", bin.Name, material.Name, expected)
		}
	}
}

func TestThrowAwayJSON(t *testing.T) {
	product := Product{EAN: "7613034383808"}
	pkg, err := NewProductPackage(product, packageDB)
	if err != nil {
		t.Fatal(err)
	}
	m1 := Material{ID: 1, Name: "Boîte carton"}
	m2 := Material{ID: 2, Name: "Film plastique"}
	m3 := Material{ID: 5, Name: "Nourriture"}
	expected, err := json.Marshal(throwAwaypackage{
		Product: ProductPackage{
			Product:   product,
			Materials: []Material{m1, m2, m3}},
		ThrowAway: map[string]string{m1.Name: "Bac à couvercle jaune", m2.Name: "Bac à couvercle vert", m3.Name: "Bac à couvercle vert"},
	})
	if err != nil {
		t.Fatal(err)
	}

	out, err := pkg.ThrowAwayJSON(packageDB)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(expected) {
		t.Errorf("ThrowAwayJson not as expected: %v != %v", string(out), string(expected))
	}
}

func TestPackageDBSet(t *testing.T) {
	if err := packageDB.Set("invalid", nil); err == nil {
		t.Error("ean should have been checked")
	} else if err != errInvalidEAN {
		t.Error(err)
	}

	ean := "9771674821123"
	if err := packageDB.Set(ean, nil); err == nil {
		t.Error("must not be able to add empty materials to a package")
	}

	m1 := Material{ID: 1, Name: "Boîte carton"}
	m2 := Material{ID: 2, Name: "Film plastique"}
	materials := []Material{m1, m1, m1, m2}
	if err := packageDB.Set(ean, materials); err != nil {
		t.Fatal(err)
	}
	pkg, err := packageDB.Get(ean)
	if err != nil {
		t.Fatal(err)
	}
	expected := 2
	if len(pkg.Materials) != expected {
		t.Errorf("Expected %v unique material, got %v", expected, len(pkg.Materials))
	}
	seenMaterials := make(map[uint]struct{}, 2)
	for _, m := range pkg.Materials {
		if _, ok := seenMaterials[m.ID]; ok {
			t.Errorf("material %v already seen once", m)
		}
		if m != m2 && m != m1 {
			t.Errorf("got %v expected %v or %v", m, m1, m2)
		}
		seenMaterials[m.ID] = struct{}{}
	}
}

var packageDB *mgoPackagesDB
var blacklistDB *mgoBlacklistDB

func TestMain(m *testing.M) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	mongoSession, err := NewMgoDB(os.Getenv("RECYCLEME_MONGO_TEST_URI"))
	if err != nil {
		logger.Fatal(err)
	}
	packageDB = NewMgoPackageDB(mongoSession, "test_")

	if err = createBins(packageDB); err != nil {
		logger.Fatal(err)
	}

	if err = createMaterials(packageDB); err != nil {
		logger.Fatal(err)
	}

	if err = createPackages(packageDB); err != nil {
		logger.Fatal(err)
	}
	blacklistDB = NewMgoBlacklistDB(mongoSession, "test_")
	ex := m.Run()
	mongoSession.Close()

	os.Exit(ex)
}
